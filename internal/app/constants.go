package app

import "time"

// Header keys
const CONTENT_TYPE = "Content-Type"
const AUTHORIZATION = "Authorization"

const BEAERER = "Bearer "

// Header Values
const APPLICATION_JSON = "application/json"
const TEXT_PLAIN = "text/plain; charset=utf-8"

// JWT expire time
const JWT_EXPIRE_TIME = 1 * time.Hour

// Time parse layout
const TIME_PARSE_LAYOUT = "2006-01-02"
