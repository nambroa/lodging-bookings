# lodging-bookings

- Built using Go 1.19
- Uses [chi router](github.com/go-chi/chi/v5) for routing.
- Uses [Alex E. SCS](github.com/alexedwards/scs/v2) for managing sessions.
- Uses [noSurf](github.com/justinas/nosurf) to prevent CSRF attacks.
- Uses [goValidator](https://github.com/asaskevich/govalidator) to sanitize strings.

## Main Goals

- A Bed & Breakfast with 2 rooms.
- Bookings & Reservations.

## Key Functionality

- Showcase the property.
- Check a room's availability.
- Allow for booking a room for 1 or more nights.
- Notification system for guests and property owners.
- (Admin) Review existing bookings, change or cancel them. Show a calendar of bookings.