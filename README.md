
# Unnamed language project

## Project directory

`language/` contains the package responsible for tokenizing and marshalling a buffer into an AST, as well as all the definitions for standard syntax
`symbols/` describes the interactions between interfaces and resolving syntax nodes into logic
`context/` contains the logic to turn a manifest into a meaningful context
`runtime/` represents all the abstractions and logic used to utilize a context as a service

## Tests to write

Interpreter test suite

:: The interpreter should be able to pass all these tests. (trying to detract from bias here)

! A BuildContext should be able to have imports
!: Should reject if the package is not a stdlib package or isn't a valid manifest file
!: Should return a cached import if it has already been added to the import cycle
   : safeguard to avoid infinite loops

! A Class should be created from an AST object node
  - including passing the object defined in 'extends' to the interface
!: Should reject if the target interface does not support creating from an object
   : does not implement `ObjectInterface`
!: Should reject if the target interface does not exist
!: Should reject if the target interface throws an error
!: Should reject if the object node has an extends statement and the selector doesn't exist
!: Should reject if any expressions in the object node require a function call

! A Class should be created from an AST method node
!: Should reject if the target interface does not support creating from a method
   : does not implement `MethodInterface`
!: Should reject if the target interface does not exist
!: Should reject if the target interface throws an error

! A ValueObject should be created from an AST object node
!: Should reject if the target interface does not support creating from an object
   : does not implement `ValueInterface`
!: Should reject if the target interface does not exist
!: Should reject if the target interface throws an error
!: Should reject if the object node has an extends property

! A Class should be able to have methods attached to it
!: Should reject if the target class does not support adding methods onto it
   : does not implement `ObjectMethodInterface`
!: Should reject if the target does not exist/is not available
!: Should reject if the target class throws an error when adding a method

! A Context should be created from a Context AST node
  - including all the classes and methods defined within

! A Context should be able to have imports and added to the object store
  - including imports that have the same domain parts (i.e. `foo.bar` and `foo.baz`)
  - including imports that have domain parts that follow a tree structure (i.e. `foo.bar` and `foo.bar.baz`)
!: Should reject if the import is not available
!: Should reject if the import has duplicate domains

! A Context should be able to evaluate a type expression
!: Should reject if the target type does not exist/is not available
!: Should reject if usage of the target type creates a usage cycle
   - 2 nodes deep
   - 3 nodes deep
   - 4 nodes deep

// todo: resource tests


### ISSUE TRACKER (ik lol)

1. 
When doing a binary expression on a nil number, the error message isn't very helpful.
```
breakdown.subtotal /* Double? = nil */ += 0.5
```
This fails with `cannot construct number from <nil>`. This should resolve with something like `cannot do binop of <nil> object`, but that would invalidate `== nil` binops. The number constructor works by providing handlers for every number type, meaning that the context of the contructed number being apart of a binop is not accessible (this is because before a binop is performed, the operand is cast to the object being operated on). Ideally this would be handled by some kind of error code, but I fear that would make a bad linkage between the type system and the number primitive.
IDK. figure it out.

2.
Right now, you can't use an index as an assignment, meaning if I wanted to effectively "map" and return a new iterable, i'd need to do this
```
iterable := someIterable
newIterable := []T{}
for (idx, val in iterable) {
   val.foo = "bar"
   newIterable = newIterable.append(val)
}
return newIterable
```
You should be able to do this:
```
for (idx in iterable) {
   iterable[idx].foo = "bar"
}
return iterable
```

3.
If you try to call a method for something that hasn't been imported yet, the error message won't be useful
```
/* without importing stickerspace.order */
stickerspace.order.FetchSalesTaxPercentage() // <- stickerspace.order.FetchSalesTaxPercentage is not callable
```
It should be `unknown selector stickerspace.order`

4.
You can use class primitives as values in a property expression and the analyzer won't fail
`fn({ product_id: String })` should fail (and does when a process is attached), but passes in analysis

5.
There's a bunch of nasty cyclic errors and nilable objects not resolving when constructing. Do `deprecated_debug()` on any complex object and you'll see what I mean. I think it's a non-issue, but GOOD HELL ITS GIVING ME A MIGRAINE

6. `break` and `continue` statements will throw an error if the statement is not declared in a loop block
```
for (...) {
   if (something) {
      break  // <-- this will fail
   }
   break  // <-- this wont fail
}
```