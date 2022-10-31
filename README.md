# lodging-bookings

- Built using Go 1.19
- Uses [chi router](github.com/go-chi/chi/v5) for routing.
- Uses [Alex E. SCS](github.com/alexedwards/scs/v2) for managing sessions.
- Uses [noSurf](github.com/justinas/nosurf) to prevent CSRF attacks.
- Uses [goValidator](https://github.com/asaskevich/govalidator) to sanitize strings.
- Uses [sodaPop](https://github.com/gobuffalo/pop?utm_source=godoc) for managing DB migrations.
- Uses [postgreSQL](https://www.postgresql.org/download/) as the DB Engine. (optional but recommended DBeaver)
- Uses [go-simple-mail](https://github.com/xhit/go-simple-mail) to send email confirmation to the customer and owner of
the cabin complex.
- Uses [RoyalUI](https://github.com/BootstrapDash/RoyalUI-Free-Bootstrap-Admin-Template) for the admin dashboard.

## Main Goals

- A Bed & Breakfast with 2 rooms.
- Bookings & Reservations.

## Key Functionality

- Showcase the property.
- Check a room's availability.
- Allow for booking a room for 1 or more nights.
- Notification system for guests and property owners.
- (Admin) Review existing bookings, change or cancel them. Show a calendar of bookings.

## Screenshots

### Main Page
![front_page](https://user-images.githubusercontent.com/42326233/199077353-731e0653-adbf-43a3-9f62-445521b2d2e3.png)

### Admin Reservations Management
![new_reservations_admin](https://user-images.githubusercontent.com/42326233/199077372-3aa770c5-ae12-4434-9b2e-c5d470c34b12.png)

### Reservations Calendar
![reservation_calendar_admin](https://user-images.githubusercontent.com/42326233/199077388-c31ac647-02c7-4707-9322-2ab6a3fa1788.png)

## How to Install

### DB

- Install Soda if it doesn't get installed via the gomod: https://gobuffalo.io/documentation/database/soda/
- Create a `database.yml` in the root directory of the project. A `database.yml.example` is included, just fill in your
postgres user and password.
- Run `soda create` to create the `lodging-bookings` database.
- Run `soda migrate` to execute the migrations and populate the database with the appropriate tables.


### MailServer

- To enable mail notifications, download and install MailHog here: https://github.com/mailhog/MailHog
- Run the mailserver before attempting to make a reservation.
- Note: MailHog is not required for the app to function. You can still make reservations, it will just not send the email
notification.
