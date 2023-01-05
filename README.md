
# split-go-ast

This is a split/extract utility for go compiler AST ir dumps.
Given an input file generated from a "-W=2" invocation of the
Go compiler, it will pick out specific functions or phases. 

Example:

```
$ cat > tiny.go << EOF
package tiny
func foo() int { return 42 }
func bar() int { return -42 }
EOF
$ go build -gcflags=-W=2 tiny.go 1> err.txt 2>&1
$ split-go-ast -func=bar -phase=escape -i=err.txt 
before escape bar
.   DCLFUNC tiny.bar ABI:ABIInternal ABIRefs:{ABIInternal} InlinabilityChecked FUNC-func() int tc(1) # tiny.go:3:6
.   DCLFUNC-Dcl
.   .   NAME-tiny.~r0 Class:PPARAMOUT Offset:0 OnStack Used int tc(1) # tiny.go:3:12
.   DCLFUNC-body
.   .   RETURN tc(1) # tiny.go:3:18
.   .   RETURN-Results
.   .   .   LITERAL--42 int tc(1) # tiny.go:3:25
$
```

