## 1. Hyper Language Data Types & Values

A _language type_ represents the behavior a _scope value_ should inherit. A _scope value_ represents an independent value that is meant to be accessed in scope and is characterized by a _language type_.

### 1.1 Language Types

#### 1.1.1 `Object` Type

An Object represents an ambiguous object that contains immutable and unsorted properties. Each object has an internal get method that takes in a key value and returns the scope value that is associated with that property.

#### 1.1.2 `ValueObject` Type

A ValueObject represents a stateful, serializable value that is described by a _Class_.

#### 1.1.3 `Callable` Type

A Callable represents an executable routine that accepts an argument list (a list of value objects) and returns either a [normal completion (annotation needed)] with a value object or a [throw completion (annotation needed)].

#### 1.1.4 `Class` Type

A Class is the canonical definition of a _ValueObject_. It serves as the definition of the interactions between different value objects which are defined by any number of class descriptors.

Class descriptors are used to define rules and isolate errors in semantic analysis, and also as the logic for mutating values in the [host environment (annotation needed)].

Classes exist as language types since they must be invoked in-scope to provide access to the constructors.

- A _class method_ is a function that pertains to an interaction of a class whos argument list begins with a reference to the subject value object.

##### 1.1.4.1 `Constructors` Class Descriptor

The constructors descriptor is a set of [constructors (annotation needed)] that establishes the logic of instanceating the parent class from another class.

--Define constructor signature (fromClass, toClass, (ValueObject\<fromClass>) => ValueObject\<toClass>)--

##### 1.1.4.2 `Operators` Class Descriptor

The operators descriptor is a set of operator methods used in evaluating certain [binary expressions (annotation needed)].

<!--this assumes that all binary expressions should assume leftOperandClass, but that might not always be the case. for instance in JS `true && "some value"` yields `"some value"`. might be useful logic in some cases-->

- An _operator method_ associates a function to a binary operator and two different classes. The method returns a new value object with the class of the left most operand and with state that corresponds to the desired result of the expression.

--Define operator method signature (leftOperandClass, rightOperandClass, operator, (ValueObject\<leftOperandClass>, ValueObject\<rightOperandClass>) => ValueObject\<leftOperandClass>)--

##### 1.1.4.3 `Comparators` Class Desciptor

The comparators descriptor is a set of comparator methods used when evaluating certain [binary expressions (annotation needed)].

- A _comparator method_ associates a function to a comparator operator and two different classes. The method returns a `Boolean` reflective of the desired result of the expression.

--Define comparator method signature (leftOperandClass, rightOperandClass, operator, (ValueObject\<leftOperandClass>, ValueObject\<rightOperandClass>) => Boolean)--

##### 1.1.4.4 `Properties` Class Descriptor

The properties descriptor is used to describe an ordered collection of properties that should be associated to a value object. Each property is described by it's key value, a getter method, and a setter method in the case that the property is mutable.

- A _getter method_ is a [class method] which is called with an empty argument list and returns the value object that is associated with a property.
- A _setter method_ is a [class method] which is called with a one-item argument list containing the new value object for the property being updated. The effect of this method should have an effect on the value returned by subsequent calls to the properties getter method.

Properties will be superceded by [prototype methods] defined in the _Prototype_ Class Descriptor that share the same key value.

##### 1.1.4.5 `Enumerable` Class Descriptor

The enumerable descriptor contains a set of zero or more [class methods] that are invoked when being evaluted in certain [enumeration expressions (annotation needed)].

**Table 1: Enumeration Methods**

| Internal Method | Signature                                                  | Description | Expression                                |
| --------------- | ---------------------------------------------------------- | ----------- | ----------------------------------------- |
| GetLength       | (ValueObject) => length                                    |             | len(ValueObject)                          |
| GetIndex        | (ValueObject, index) => ValueObject                        |             | ValueObject[index]                        |
| SetIndex        | (ValueObject, index, newValue) => nil                      |             | ValueObject[index] = newValue             |
| GetRange        | (ValueObject, fromIndex, toIndex) => ValueObject           |             | ValueObject[fromIndex:toIndex]            |
| SetRange        | (ValueObject, fromIndex, toIndex, newValue) => ValueObject |             | ValueObject[fromIndex:toIndex] = newValue |

##### 1.1.4.6 `Prototype` Class Descriptor

The prototype descriptor defines a collection of key values that associate read-only [class methods] that will be accessible as properties from the value object.

The inclusion of a prototype descriptor doesn't imply prototype-based inheritance for a value object in the sense that the methods exist within scope of the object in the [host environment (annotation needed)]. Rather, expressions that target a value object's properties are evaluated by looking up the class prototype methods first, then by any properties defined in the _Properties_ class descriptor.

Prototype methods do not exist as accessible properties on the class itself, only on the value object.

<!-- TODO: write these
### 1.2 Language Primitives

#### 1.2.1 `nil` Primitive

#### 1.2.2 `Boolean` Primitive

#### 1.2.3 `String` Primitive

#### 1.2.4 Numeric Primitives

##### 1.2.4.1 `Number` Primitive

##### 1.2.4.2 `Double` Primitive

##### 1.2.4.3 `Float` Primitive

##### 1.2.4.4 `Integer` Primitive

#### 1.2.5 `DateTime` Primitive

#### 1.2.6 `Map` Primitive

### 1.3 Language

#### 1.3.1 `Nilable`

#### 1.3.2 `Array`

#### 1.3.3 `Map`

#### 1.3.4 `Error`

#### 1.3.5 `Function`

##### 1.3.5.1 Additional scope

##### 1.3.5.2 `guard` Directive -->
