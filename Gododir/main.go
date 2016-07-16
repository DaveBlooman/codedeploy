package main

import do "gopkg.in/godo.v2"

func tasks(p *do.Project) {
	p.Task("server", nil, func(c *do.Context) {
		// rebuilds and restarts when a watched file changes
		c.Start("main.go", do.M{"$in": "./"})
	}).Src("*.go", "**/*.go").
		Debounce(3000)

	p.Task("save-deps", nil, func(c *do.Context) {
		c.Bash("godep save")
	})

	p.Task("build", do.S{"save-deps"}, func(c *do.Context) {
		c.Bash("go build -o mux-api")
	})
}

func main() {
	do.Godo(tasks)
}
