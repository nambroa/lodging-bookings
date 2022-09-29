- All the files for a package MUST exist on the same directory. If you have a folder called render with a file called
  render1.go and it's part of the render package, you cannot have a render2.go file on the directory handlers
  saying `package render`, it will not be able to find functions that live in render1.go (and vice versa)