# Go Design Patterns

Project/Repository designs for applications written in go.

## [StepKit Pattern](/stepkit)

An implementation of Vertical Slice Architecture where **_each unit of work is a self-contained_** vertical slice containing everything it needs (handler + interfaces + implementations), instead of horizontal layers shared across the whole app.
