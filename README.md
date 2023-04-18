# ![RealWorld Example App](logo.png)

> ### Hexgonal architecture codebase written in Go and containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.


### [Demo](https://demo.realworld.io/)&nbsp;&nbsp;&nbsp;&nbsp;[RealWorld](https://github.com/gothinkster/realworld)


This codebase was created to demonstrate a fully fledged backend application built with Go using hexagonal architecture, including CRUD operations, authentication, routing, pagination, and more.

For more information on how to this works with other frontends/backends, head over to the [RealWorld](https://github.com/gothinkster/realworld) repo.


# How it works

This application demonstrates the use of hexagonal archtitecture in Go. All
components are rigorously decoupled from one other, and dependencies only point
inwards, towards the domain models for any given route.

For a primer on hexagonal architecture, see this blog entry from
[Netflix](https://netflixtechblog.com/ready-for-changes-with-hexagonal-architecture-b315ec967749).

The business logic, stored under `internal/service` is agnostic to the
controllers that receive requests from the outside world. I've used the [Fiber
package](https://gofiber.io/) as my web server, but I could just as well have
used the standard library http package or Gin. Check out the web server code
under `internal/controller/rest`.

Neither does the business logic care about the nature of its datastore. This
example uses SQLite (`internal/repository/sqlite`), but thanks to the Repository
pattern, it would be trivial to use Postgres, MongoDB, or a human typing at a
terminal instead.

We can go further still. By embedding a `Presenter` interface in our REST
handlers, we decouple the formatting of responses from the logic that
coordinates the parsing of the request and invocation of the business logic. As
you'll see, this makes testing handlers trivial, since the `Presenter` becomes
solely responsible for rendering responses. Take a look at `internal/presenter`
for more details.

# Getting started

1. Copy `.env_template` to `.env` and add an RSA private key in the space
   indicated. `.env` is gitignored. Be sure to keep it that way.
2. Load `.env` into your environment however you please. Here's my personal preference:
    ```bash
    $ set -o allexport; source ./.env; set +o allexport
    ```
3. Run the app interactively with `make docker_run_it`, or `make docker_run` to
   run it in the background.
4. Test the app with the [RealWorld Postman collection](https://github.com/gothinkster/realworld/blob/main/api/Conduit.postman_collection.json).

# Progress

Here's what's been implemented so far:

[x] Users
[x] Authentication
[ ] Profiles
[ ] Articles
[ ] Comments
[ ] Tags