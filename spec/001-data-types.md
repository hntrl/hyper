## 1. Hyper Language Data Types & Values

A _language type_ represents the behavior a _language value_ should inherit. A _language value_ represents an independent value that is meant to be accessed in scope and is characterized by a _language type_.

### 1.1 Language Types

#### 1.1.1 `Object` Type

An Object is a collection of immutable and unsorted properties. Each property is either a data property, or an accessor property.

- A _data property_ associates a key value to a _language value_.
- An _accessor property_ associates a key value with a getter function. The _getter function_ is used to retrieve a _language value_ that is associated with a property.

#### 1.1.2 `ValueObject` Type

A ValueObject represents a stateful, serializable value that is described by a _Class_.

Value objects have certain logic handler requirements when it comes to how the value objects class is defined ((see [Intrinsic Class Logic Relationships (annotation needed)] and [ValidateClassRelationship(C, V) (annotation needed)])).

#### 1.1.3 `Class` Type

A Class is the canonical definition of a _ValueObject_. It serves as the definition of the interactions between different value objects which are defined by any number of class descriptors.

Class descriptors are used to define rules and isolate errors in semantic analysis, and also as the logic for mutating values in the [host environment (annotation needed)].

##### 1.1.3.1 `Constructors` Class Descriptor

The constructors descriptor is a set of [constructors (annotation needed)] that establishes the logic of instanceating the parent class from another class.

--Define constructor signature (fromClass, toClass, (ValueObject<fromClass>) => ValueObject<toClass>)--

##### 1.1.3.2 `Operations` Class Descriptor

The operations descriptor is a set of operator methods used in evaluating certain [binary expressions (annotation needed)].

<!--this assumes that all binary expressions should assume leftOperandClass, but that might not always be the case. for instance in JS `true && "some value"` yields `"some value"`. might be useful logic in some cases-->

- An _operator method_ associates a function to a binary operator and two different classes. The method returns a new value object with the class of the left most operand and with state that corresponds to the desired result of the expression.

--Define operator method signature (leftOperandClass, rightOperandClass, operator, (ValueObject<leftOperandClass>, ValueObject<rightOperandClass>) => ValueObject<leftOperandClass>)--

##### 1.1.3.3 `Comparators` Class Desciptor

The comparators descriptor is a set of comparator methods used when evaluating certain [binary expressions (annotation needed)].

- A _comparator method_ associates a function to a comparator operator and two different classes. The method returns a `Boolean` reflective of the desired result of the expression.

--Define comparator method signature (leftOperandClass, rightOperandClass, operator, (ValueObject<leftOperandClass>, ValueObject<rightOperandClass>) => Boolean)--

##### 1.1.3.4 `Properties` Class Descriptor

The properties descriptor is used to describe the collection of properties that should be associated to a value object. Each property is described by it's key value, the class the property will hold, and whether or not the property is mutable.

Properties will be superceded by prototype [methods (annotation needed)] defined in the _Prototype_ Class Descriptor that share the same key value.

##### 1.1.3.5 `Enumerable` Class Descriptor

The enumerable descriptor contains a set of zero or more [methods (annotation needed)] that are invoked when being evaluted in certain expressions.

**Table 1: Enumeration Methods**

| Internal Method | Signature                                                  | Description | Expression                                |
| --------------- | ---------------------------------------------------------- | ----------- | ----------------------------------------- |
| GetIndex        | (ValueObject, index) => ValueObject                        |             | ValueObject[index]                        |
| SetIndex        | (ValueObject, index, newValue) => ValueObject              |             | ValueObject[index] = newValue             |
| GetRange        | (ValueObject, fromIndex, toIndex) => ValueObject           |             | ValueObject[fromIndex:toIndex]            |
| SetRange        | (ValueObject, fromIndex, toIndex, newValue) => ValueObject |             | ValueObject[fromIndex:toIndex] = newValue |

##### 1.1.3.6 `Prototype` Class Descriptor

The prototype descriptor defines a collection of [methods (annotation needed)] that will be accessible from the value object. Each method can return any class and will be called with an argument list with the sole argument being the current value object.

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

#### 1.3.1 `NilableObject`

#### 1.3.2 `Iterable`

#### 1.3.3 `Function` -->
