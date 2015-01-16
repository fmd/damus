package main

type Chain []string

func (c Chain) Get(link string) Chain {
	p := -1
	for i, l := range c {
		if l == link {
			p = i
			break
		}
	}

	if p == -1 {
		return Chain{}
	}

	return c[p:]
}

func (c Chain) Final() Chain {
	return c[len(c)-1:]
}

func (c Chain) String() string {
	return c[0]
}
