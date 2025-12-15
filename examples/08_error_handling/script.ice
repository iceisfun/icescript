
func fail(msg) {
    panic(msg)
}

func nested() {
    fail("something went wrong")
}

func main() {
    print("Starting...")
    nested()
    print("Should not be reached")
}

main()
