package collisions

// Sensor marks an entity whose collisions are detected — Hit confirm/expire
// behaves normally — but never physically resolved: no MTV push-apart, no
// CollisionHandler dispatch, for any contact it's part of.
type Sensor struct{}
