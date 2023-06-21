## 1. Contexts & Interfaces

Conceptually, the main purpose of hyper is to provide a way to describe the interactions between different parts of a system in a way that is both human-readable and machine-readable. In order to create that separation, hyper systems are architected around the broader [domain (annotation needed)] being divided up into sections that are concerned with their own independent vertical, called a _context_.

Contexts are described by a set of _context items_, which represent the different elements of a context. Each context item when being defined requires that the item be defined with an interface. An _interface_ describes the behavior of a context item, and isolates how context items should be interpreted in the [host environment (annotation needed)]. Interfaces are delimited as either a _context object_ or a _context method_.

- A _context object_ represents an item that is meant to represent a classification of a stateful value.
- A _context method_ represents an item that is meant to represent a stateless routine (meaning that state is not shared among different methods).

Context objects can be further described by a _context object method_, which represents a routine that has some sort of association with the item's interface.

Contexts for all intensive purposes represent their own services, and are meant to be deployed as such. The only means of interaction between between multiple contexts is through a [message broker (annotation needed)].

<!-- TODO: decide if standard interfaces belong here; i don't know if it makes more sense to isolate this from the spec since those interfaces rely on things like stdlib which makes bad linkage between standard interfaces and the concept of interfaces as a whole -->

### 1.1 Standard Interfaces

#### 1.1.1 `enum` Interface

The enum interface object is used to describe a `Class` whos value is restricted to the values defined within the enum. Enums are described by any number of unique [enum fields (annotation needed)].

```
enum Color {
  Red   "red"
  Green "green"
  Blue  "blue"
}
```

#### 1.1.2 `type` Interface

The type interface object is used to describe a `Class` whos value is mutable and holds the properties that are defined within the type. Types are described by any number of unique [type fields (annotation needed)].

```
type Person {
  name String
  age  Number
}
```

#### 1.1.3 `parameter` Interface

The parameter interface object is used to describe a value that is passed into the context. Parameters are described by type which is delimited by a [type field (annotation needed)], as well as an optional name, description, and defaultValue (which must have the same type as described in the type field). Parameters should be used to define values that might not have a place in defining the context, but are still needed in the deployment (like secret keys).

```
parameter SecretKey {
  type         String
  name         = "key"
  description  = "My super secret key"
  defaultValue = "password1234"
}
```

#### 1.1.4 `template` Interface

--TODO: template isn't concrete yet. it relies on a third party library which is probably going to change (see internal/interfaces/template.go)--

#### 1.1.5 `file` Interface

The file interface object is used to describe a type of file that should be stored in a file-store. Files are described by an array of allowed [mime-types (annotation needed)] which are described in a [field assignment (annotation needed)], as well as an optional name and description. The File Class and ValueObject have methods related to managing the state of a file instance.

--TODO: write what aforementioned methods are--

```
file Image {
  allowed = []mime.MimeType{"image/png", "image/jpeg"}
}
```

#### 1.1.6 State Interfaces

##### 1.1.6.1 `entity` Interface

```
entity Person {
  name String
  age  Number
}
func (Person) onCreate(p: Person) {

}
func (Person) onUpdate(p: Person) {

}
func (Person) onDelete(p: Person) {

}
```

##### 1.1.6.2 `projection` Interface

```
projection PersonDetails {
  name String
  age  Number
}
func (PersonDetails) onEvent(event: PersonCreated) {

}
```

#### 1.1.7 Stream Interfaces

##### 1.1.7.1 `command` Interface

```
command CreatePerson(p: Person) Person {}
```

##### 1.1.7.2 `event` Interface

```
event PersonCreated {
  subject Person
}
```

##### 1.1.7.3 `query` Interface

```
query GetPerson(id: String) Person {}
```

##### 1.1.7.4 `sub` Interface

```
sub PersonCreatedSubscription(ev: PersonCreated) {}
```

#### 1.1.8 Access Interfaces

##### 1.1.8.1 `grant` Interface

The grant interface object is used to describe a permission and works as a means of controlling access to a resource. Grants are described by a name and a description.

```
grant ManageResource {
  name = "Manage Resources"
  description = "Grants access to manage resources"
}
```
