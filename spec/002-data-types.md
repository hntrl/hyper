## 2. Hyper Language Data Interfaces & Values

A _scope value_ represents an independent value that is meant to be accessed in scope and is characterized by a _language interface_.

### 2.1 Language Types

#### 2.1.1 `Object` Type

An Object represents an ambiguous object that contains immutable and unsorted properties. Each object has an internal get method that takes in a key value and returns the scope value that is associated with that property.

#### 2.1.2 `ValueObject` Type

A ValueObject represents a stateful, serializable value that is described by a _Class_.

#### 2.1.3 `Callable` Type

A Callable represents an executable routine that accepts an argument list (a list of value objects) and returns either a [normal completion (annotation needed)] with a value object or a [throw completion (annotation needed)] with an Error object.

#### 2.1.4 `Class` Type

A Class is the canonical definition of a _ValueObject_. It serves as the definition of the interactions between different value objects which are defined by any number of class descriptors.

Class descriptors are used to define rules and isolate errors in semantic analysis, and also serve as the logic for mutating values in the [host environment (annotation needed)].

Classes exist as language interfaces since they must be invoked in-scope to provide access to the constructors.

- A _class method_ is a function that pertains to an interaction of a class whos argument list begins with a reference to the subject value object.

##### 2.1.4.1 `Constructors` Class Descriptor

The constructors descriptor is a set of [constructors (annotation needed)] that establishes the logic of instanceating the parent class from another class.

--Define constructor signature (fromClass, toClass, (ValueObject\<fromClass>) => ValueObject\<toClass>)--

##### 2.1.4.2 `Operators` Class Descriptor

The operators descriptor is a set of operator methods used in evaluating certain [binary expressions (annotation needed)].

<!--this assumes that all binary expressions should assume leftOperandClass, but that might not always be the case. for instance in JS `true && "some value"` yields `"some value"`. might be useful logic in some cases-->

- An _operator method_ associates a function to a binary operator, a target class, and an operand class. The method returns a value object assuming the target class and with state that corresponds to the desired result of the expression.

--Define operator method signature (leftOperandClass, rightOperandClass, operator, (ValueObject\<leftOperandClass>, ValueObject\<rightOperandClass>) => ValueObject\<leftOperandClass>)--

##### 2.1.4.3 `Comparators` Class Desciptor

The comparators descriptor is a set of comparator methods used when evaluating certain [binary expressions (annotation needed)].

- A _comparator method_ associates a function to a comparator operator and two different classes. The method returns a `Boolean` reflective of the desired result of the expression.

--Define comparator method signature (leftOperandClass, rightOperandClass, operator, (ValueObject\<leftOperandClass>, ValueObject\<rightOperandClass>) => Boolean)--

##### 2.1.4.4 `Properties` Class Descriptor

The properties descriptor is used to describe an ordered collection of properties that should be associated to a value object. Each property is described by it's key value, the class the property assumes, a getter method in the case that the property is accessible, and a setter method in the case that the property is mutable.

- A _getter method_ is a [class method] which is called with an empty argument list and returns the value object that is associated with a property.
- A _setter method_ is a [class method] which is called with a one-item argument list containing the new value object for the property being updated. The effect of this method should have an effect on the value returned by subsequent calls to the properties getter method.

Properties will be superceded by [prototype methods] defined in the _Prototype_ Class Descriptor that share the same key value.

##### 2.1.4.5 `Enumerable` Class Descriptor

The enumerable descriptor contains a set of zero or more [class methods] that are invoked when being evaluted in certain [enumeration expressions (annotation needed)].

**Table 1: Enumeration Methods**

| Internal Method | Signature                                                  | Description | Expression                                |
| --------------- | ---------------------------------------------------------- | ----------- | ----------------------------------------- |
| GetLength       | (ValueObject) => length                                    |             | len(ValueObject)                          |
| GetIndex        | (ValueObject, index) => ValueObject                        |             | ValueObject[index]                        |
| SetIndex        | (ValueObject, index, newValue) => nil                      |             | ValueObject[index] = newValue             |
| GetRange        | (ValueObject, fromIndex, toIndex) => ValueObject           |             | ValueObject[fromIndex:toIndex]            |
| SetRange        | (ValueObject, fromIndex, toIndex, newValue) => ValueObject |             | ValueObject[fromIndex:toIndex] = newValue |

##### 2.1.4.6 `Prototype` Class Descriptor

The prototype descriptor defines a collection of key values that associate read-only [class methods] that will be accessible as properties from the value object.

The inclusion of a prototype descriptor doesn't imply prototype-based inheritance for a value object in the sense that the methods exist within scope of the object in the [host environment (annotation needed)]. Rather, expressions that target a value object's selector are evaluated by looking up the class prototype methods first, then by any properties defined in the _Properties_ class descriptor.

Prototype methods do not exist as accessible properties on the class itself, only on the value object assumed by the class.

##### 2.1.4.7 `ClassProperties` Class Descriptor

The class properties descriptor defines a collection of key values that associate read-only [scope values] that will be accessible as properties on the uninstanceated class object.

### 2.2 Built-in Language Objects

#### 2.2.1 `Nil` Object

The Nil type has exactly one value that represents the lack of an object, called _nil_.

#### 2.2.2 `Boolean` Object

A boolean type represents a logical truth value denoted by the predeclare constants `true` and `false`.

#### 2.2.3 `String` Object

A string type represents a (possibly empty) sequence of bytes. The number of bytes is called the length of the string and is never negative.

#### 2.2.4 Numeric Objects

Any type that represents a numeric value are called numeric types. All numeric types are constructable, operable, and comparable to all other numeric types. The predeclared numeric types are:

| Object    | Description                                                         |
| --------- | ------------------------------------------------------------------- |
| `Number`  | Represents a floating-point number                                  |
| `Double`  | Represents a rounded-up floating-point number with a precision of 2 |
| `Float`   | Analogous to `Number`                                               |
| `Integer` | Represents a signed integer                                         |

#### 2.2.5 `Map` Object

A map represents a generic collection of key-value pairs that doesn't conform to a specific set of properties. Keys and values can be mutated so long as the updated value can be constructed into the previous value. Keys are unique, unordered, and indexed by a string. The number of elements is the length of the map and is never negative.

#### 2.2.6 `Nilable` Object

A nilable represents a value that can be assumed to have a nil value or a value of a specific type. The Nilable type is a generic type that takes in a class parameter that represents the class of the value that is assumed (the parent class) when the Nilable is not nil. The Nilable descriptors share the same descriptors as the class but act as a wrapper to allow for nil values to be handled.

#### 2.2.7 `Array` Object

An array is a numbered sequence of elements of a single class, called the item class. The number of elements is called the length of the array and is never negative.

#### 2.2.8 `Error` Object

An error represents a value that is thrown when an error occurs in the execution of a routine. An error is described by a name string, a message string, and a generic data value to annotate the specific details of the error.

#### 2.2.9 `Function` Object

A function represents a sub-routine to be called through an interface so the control flow can interact with logic in the [host environment (annotation needed)]. A function implements `Callable` and is described by a list of arguments classes, and a return class.
