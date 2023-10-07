# Compile

From the top-level directory run

```
tinygo build -size short -o firmware.uf2 -target=pico zs-042/main.go
```

# Compile and Flas

```
tinygo flash -target=pico zs-042/main.go
```

# Monitor

```
tinygo monitor
```