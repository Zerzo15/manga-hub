package main

func main() {
	client := NewAPIClient("http://localhost:8080")
	ui := NewTUI(client, "localhost:9090")
	if err := ui.Run(); err != nil {
		panic(err)
	}
}
