package appcontext

// required to avoid dependency cycle storage <-> app
type Key string 
const KeyUserID Key = "userID"