# lodging-bookings

- Built using Go 1.19
- Uses [chi router](github.com/go-chi/chi/v5) for routing.
- Uses [Alex E. SCS](github.com/alexedwards/scs/v2) for managing sessions.
- Uses [noSurf](github.com/justinas/nosurf) to prevent CSRF attacks.
- Uses [goValidator](https://github.com/asaskevich/govalidator) to sanitize strings.
- Uses [sodaPop](https://github.com/gobuffalo/pop?utm_source=godoc) for managing DB migrations.
- Uses [postgreSQL](https://www.postgresql.org/download/) as the DB Engine. (optional but recommended DBeaver)

## Main Goals

- A Bed & Breakfast with 2 rooms.
- Bookings & Reservations.

## Key Functionality

- Showcase the property.
- Check a room's availability.
- Allow for booking a room for 1 or more nights.
- Notification system for guests and property owners.
- (Admin) Review existing bookings, change or cancel them. Show a calendar of bookings.

## How to Install

### DB

- Install Soda if it doesn't get installed via the gomod: https://gobuffalo.io/documentation/database/soda/
- Create a `database.yml` in the root directory of the project. A `database.yml.example` is included, just fill in your
postgres user and password.
- Run `soda create` to create the `lodging-bookings` database.
- Run `soda migrate` to execute the migrations and populate the database with the appropriate tables.
