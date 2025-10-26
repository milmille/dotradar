# DotRadar

An interactive weather radar in your terminal

## Usage
Start DotRadar, centered on a state of your choice
```
dotradar -s minnesota
```

## Keybinds

h   - pan left

j   - pan down

k   - pan up

l   - pan right

u   - zoom out

d   - zoom in

ESC - quit

## Build From Source

Requirements: go 1.23 or newer

```
git clone https://github.com/milmille/dotradar.git && cd dotradar
```

build
```
go build . -o dotradar
```

[!NOTE]

DotRadar uses [Unicode Octant](https://www.unicode.org/charts/PDF/Unicode-16.0/U160-1CC00.pdf) characters which 
may not be implemented by the font you're using, resulting in a replacement character being drawn instead. You 
can try using a font that has these characters implemented such as [Cascadia Code](https://github.com/microsoft/cascadia-code).
More character support is coming soon.

