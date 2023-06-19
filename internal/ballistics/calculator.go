package ballistics

import (
	"math"

	"go.pitz.tech/units/length"
)

// Calculator defines a ballistic calculator that computes common values for ballistics such as MuzzleVelocity (or u),
// ProjectileEnergy, MaxDistance, and even plot a Trajectory.
type Calculator struct{}

// EstimateDistance implements a common equation used by snipers to estimate distances to target using a mil-dot reticle
// in first focal plane scopes.
func (c Calculator) EstimateDistance(targetSize length.Length, mils float64) length.Length {
	distanceToTarget := float64(targetSize) * 1000 / mils

	return length.Length(distanceToTarget)
}

// ProjectileAcceleration computes the velocity of a projectile given the pressure applied to the projectile, its
// diameter, and weight.
func (c Calculator) ProjectileAcceleration(pressure, diameter, mass float64) float64 {
	return pressure * math.Pi * math.Pow(diameter/2, 2) / mass
}

// MuzzleVelocity computes a projectiles velocity given the pressure, diameter, and length of the barrel along with
// the weight of the projectile.
func (c Calculator) MuzzleVelocity(acceleration, length float64) float64 {
	return math.Sqrt(2 * acceleration * length)
}

// ProjectileEnergy computes the energy of a projectile given its velocity and weight.
func (c Calculator) ProjectileEnergy(velocity, mass float64) float64 {
	return mass * math.Pow(velocity, 2) / 2
}

func (c Calculator) Trajectory(g, v0, theta float64, dragFunc func(v float64) float64) func(x int) (t, y, velocity float64) {
	//area := math.Pi * math.Pow(diameter/2, 2)
	sinTheta := math.Sin(theta)
	cosTheta := math.Cos(theta)

	v := v0
	ux := v0 * cosTheta
	uy := v0 * sinTheta

	lastX := 0
	y := 0.0 // -height over bore
	t := 0.0

	return func(x int) (_, _, _ float64) {
		deltaT := float64(x-lastX) / ux
		lastX = x

		drag := dragFunc(v)

		y = y + (uy * deltaT)

		ux = ux - drag*deltaT
		uy = uy - (drag+g)*deltaT

		v = math.Sqrt(math.Pow(ux, 2) + math.Pow(uy, 2))

		t += deltaT

		return t, y, v
	}
}
