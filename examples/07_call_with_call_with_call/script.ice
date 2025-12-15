func makeAdder(x) {
    return func(y) {
        return x + y
    }
}

var add1 = makeAdder(1)
var add2 = makeAdder(2)

print(add2(add1(add2(add1(10)))))
