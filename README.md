
# Unnamed language project

## Project directory

`language/` contains the package responsible for tokenizing and marshalling a buffer into an AST, as well as all the definitions for standard syntax
`resource/`

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

