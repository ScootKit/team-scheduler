/* utils for getting user's location */

// Disabled for the WannPassts internal build: upstream fetched the user's IP/location from the
// third-party service geolocation-db.com (a privacy leak). We do not call any external geolocation
// service; callers handle a null result.
export const getLocation = () => {
  return Promise.resolve(null)
}
