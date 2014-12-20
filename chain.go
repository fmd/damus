package main

import (
    "errors"
    "fmt"
)

type Chain []string

func (c Chain) Get(link string) (Chain, error) {
    p := -1
    for i, l := range c {
        if l == link {
            p = i
            break
        }
    }

    if p == -1 {
        return nil, errors.New(fmt.Sprintf("Could not find build `%s` in chain.", link))
    }

    return c[p:], nil
}

func (c Chain) Final() Chain {
    return c[len(c)-1:]
}

func (c Chain) String() string {
  return c[0]
}
