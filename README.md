# fipple

[fipple](https://fipress.org/project/fipple) is a RESTful http framework in Go.

**Usage**

1. Register service
```
fipple.Get("/",function(ctx fipple.Context) {
    //a GET 
})

```


2. Start
```
fipple.Start(port)
```