package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Tester struct {
	Client           *docker.Client
	Image            string
	Build            string
	Stamp            string
	LogDir           string
	InspectFrequency int
}

func (t *Tester) FullName() string {
	return fmt.Sprintf("%s-%s", t.Image, t.Build)
}

func TimeHash() string {
	h := fnv.New64a()
	defer h.Reset()
	h.Write([]byte(time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))
}

func NewTester(endpoint string, image string, build string, logDir string, freq int) (*Tester, error) {
	c, err := docker.NewClient(endpoint)
	if err != nil {
		return nil, err
	}

	stamp := TimeHash()

	t := &Tester{
		Client:           c,
		Image:            image,
		Build:            build,
		Stamp:            stamp,
		LogDir:           logDir,
		InspectFrequency: freq,
	}

	//Create logs directory
	t.LogDir = filepath.Join(logDir, image, stamp)
	err = os.MkdirAll(t.LogDir, 0755)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	log.WithFields(log.Fields{
		"image":  t.Image,
		"build":  t.Build,
		"stamp":  t.Stamp,
		"logdir": t.LogDir,
	}).Info("Damus has begun testing.")

	return t, nil
}

func (t *Tester) Check(s *Test) (int, error) {
	for {
		c, err := t.Client.InspectContainer(s.Id)
		if err != nil {
			return 1, err
		}

		if c.State.ExitCode != 0 {
			return c.State.ExitCode, nil
		}

		if !c.State.Running {
			return 0, nil
		}

		time.Sleep(time.Duration(t.InspectFrequency) * time.Millisecond)
	}
}

func (t *Tester) SaveImage(s *Test) error {
	opts := docker.CommitContainerOptions{
		Container:  s.Id,
		Repository: t.FullName(),
		Tag:        "latest",
	}
	_, err := t.Client.CommitContainer(opts)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tester) Commit(s *Test) error {
	err := t.Create(s)
	if err != nil {
		return err
	}

	err = t.Start(s)
	if err != nil {
		return err
	}

	code, err := t.Check(s)
	if code != 0 {
		return errors.New("Commit action quit with non-zero exit code. Aborting.")
	}

	err = t.SaveImage(s)
	if err != nil {
		return err
	}

	return nil
}

type TestResult struct {
	Test  *Test
	Code  int
	Error error
}

func (t *Tester) StartTest(test *Test, results chan TestResult) {
	go func(test *Test, results chan TestResult) {
		log.WithFields(log.Fields{
			"test": test.Name,
		}).Info("Started Test.")

		r := TestResult{
			Test:  test,
			Code:  0,
			Error: nil,
		}

		err := t.Create(test)
		if err != nil {
			r.Code = 1
			r.Error = err
			results <- r
			return
		}

		err = t.Start(test)
		if err != nil {
			r.Code = 1
			r.Error = err
			results <- r
			return
		}
		r.Code, r.Error = t.Check(test)
		results <- r
		return
	}(test, results)
}

func (t *Tester) Test(tests []Test) (int, error) {
	results := make(chan TestResult, len(tests))

	for _, test := range tests {
		s := &Test{Cmd: test.Cmd, Name: test.Name}
		t.StartTest(s, results)
	}

	finishedTests := 0

	for finishedTests != len(tests) {
		select {
		case r := <-results:
			err := t.Log(r.Test)

			log.WithFields(log.Fields{
				"test":  r.Test.Name,
				"code":  r.Code,
				"error": r.Error,
				"log":   filepath.Join(t.LogDir, r.Test.Name),
			}).Info("Finished Test.")

			if err != nil {
				return r.Code, err
			}
			if r.Code != 0 {
				return r.Code, r.Error
			}
			if r.Error != nil {
				return r.Code, r.Error
			}

			finishedTests++
		}
	}

	return 0, nil
}

type LogResult struct {
	Error error
}

func (t *Tester) Log(s *Test) error {
	var w bytes.Buffer
	writer := bufio.NewWriter(&w)
	opts := docker.LogsOptions{
		Container:    s.Id,
		OutputStream: writer,
		ErrorStream:  writer,
		Follow:       false,
		Stdout:       true,
		Stderr:       true,
		Timestamps:   true,
		Tail:         "",
		RawTerminal:  true,
	}
	err := t.Client.Logs(opts)
	if err != nil {
		return err
	}
	//Save logs to file.
	err = ioutil.WriteFile(filepath.Join(t.LogDir, s.Name), w.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tester) Create(s *Test) error {
	c, err := t.Client.CreateContainer(s.Config(t.Image, t.Build, t.Stamp))
	if err != nil {
		return err
	}

	s.Id = c.ID
	return nil
}

func (t *Tester) Start(s *Test) error {
	if len(s.Id) == 0 {
		return errors.New("A container has not been created for this test.")
	}
	return t.Client.StartContainer(s.Id, &docker.HostConfig{})
}
