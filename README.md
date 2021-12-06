# go type registry
register type and create object with string, with code generation

## Usage
for your project
```
registry_content_generator --mode distributed --path $PROJECT_PATH --package $YOUR_PACKAGE_NAME
```
will generate __type_registry.go in every subpackage, register every type in your package to registry

then import registry package, registry.New(string) will create object corresponding to type name string

for dependency

```
registry_content_generator --path $DEPENDENCY_PATH --package $DEPENDENCY_PACKAGE_NAME
```
will generate registryContent.go, register types of dependency package recursively

## Original idea
I need to
1. Serialize any data
2. Deserialize to corresponding type remotely

so I save type name in my serialized data, use type registry to recreate serialized object
