# Pharmacy booking website

## Server

### Requirement
 - MySQL
 - Redis
 - SphinxSearch

### Build

```sh
go build cmd/apteka/main.go
```

## Efarma integration
Upload products from efarma database to server

### Requirement
- Microsoft SQL

### Build
```sh
go build cmd/main.go
```

## Store client

Client for pharmacies to view bookings

### Requirement
 - Open GL

### Build

Notify:
```sh
go build -o notify.exe -ldflags -H=windowsgui notify/main.go
```

App:
```sh
fyne package
```
