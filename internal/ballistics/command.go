package ballistics

import (
	"encoding/csv"
	"fmt"
	"math"
	"strconv"

	"github.com/urfave/cli/v2"
	"go.pitz.tech/lib/flagset"
	"go.pitz.tech/units"
	"go.pitz.tech/units/gravity"
	"go.pitz.tech/units/length"
	"go.pitz.tech/units/mass"
)

type ChartTrajectoryConfig struct {
	Unit string `json:"unit" usage:"specify what units should be used when formatting the table" default:"imperial"`

	BulletCaliber     *length.Length `json:"bullet_caliber"  usage:"caliber of projectile (inches)" default:"223th"`
	BulletWeight      *mass.Mass     `json:"bullet_weight" usage:"weight of the projectile (grains)" default:"55gr"`
	CartridgePressure float64        `json:"cartridge_pressure" usage:"load pressure (PSI)" default:"55000"`
	BarrelLength      *length.Length `json:"barrel_length" usage:"length of the barrel (inches)" default:"18in"`
	ZeroDistance      float64        `json:"zero_distance" usage:"distance between your zero points (x2 - x1)"`
	DragCoefficient   float64        `json:"drag_coefficient" usage:"configure the drag coefficient for the trajectory" default:"0.2"`

	Step  int `json:"step" usage:"specify the size of the step to take between each distance marker" default:"100"`
	Range int `json:"range" usage:"specify the maximum distance marker" default:"1000"`
}

type ChartRangeConfig struct {
	Unit string `json:"unit" usage:"specify what units should be used when formatting the distance" default:"imperial"`

	Step  int `json:"step" usage:"specify the size of the step in mils taken between each value" default:"1"`
	Range int `json:"range" usage:"specify the maximum mil value" default:"30"`
}

func p[T any](val T) *T {
	return &val
}

var (
	imperialShort = units.Unit[length.Length]{
		{length.Foot, []string{"ft", "'"}},
		{length.Yard, []string{"yd"}},
	}

	siShort = units.Unit[length.Length]{
		{length.Meter, []string{"m"}},
		{length.Kilometer, []string{"km"}},
	}

	plotRange = &ChartRangeConfig{}

	plotTrajectory = &ChartTrajectoryConfig{
		BulletCaliber: p(length.Length(0.223 * float64(length.Inch))),
		BulletWeight:  p(55 * mass.Grain),
		BarrelLength:  p(18 * length.Inch),
	}

	Slug = 14593902937205 * mass.Nanogram

	Command = &cli.Command{
		Name:  "ballistics",
		Usage: "Common operations for working with ballistics.",
		Subcommands: []*cli.Command{
			{
				Name:  "range",
				Usage: "Common operations for working with ranges.",
				Subcommands: []*cli.Command{
					{
						Name:      "plot",
						Usage:     "Plot out a range estimation table based on subjects known height.",
						UsageText: "plot <heights...>",
						Flags:     flagset.ExtractPrefix("em", plotRange),
						Action: func(ctx *cli.Context) error {
							var unit units.Unit[length.Length]

							switch plotRange.Unit {
							case "imperial":
								unit = imperialShort
							case "si":
								unit = siShort
							default:
								return fmt.Errorf("unsupported unit: %s", plotRange.Unit)
							}

							calc := &Calculator{}

							out := csv.NewWriter(ctx.App.Writer)
							out.Comma = '\t'

							defer out.Flush()

							headers := ctx.Args().Slice()
							lengths := make([]length.Length, len(headers))
							for i := 0; i < len(headers); i++ {
								err := (&lengths[i]).Set(headers[i])
								if err != nil {
									return err
								}
							}

							_ = out.Write(append([]string{"mil"}, headers...))

							for mil := 1; mil <= plotRange.Range; mil += plotRange.Step {
								distances := make([]string, len(lengths))
								for i, length := range lengths {
									distances[i] = unit.Format(calc.EstimateDistance(length, float64(mil)))
								}

								_ = out.Write(append(
									[]string{strconv.Itoa(mil)},
									distances...,
								))
							}

							return nil
						},
					},
				},
			},
			{
				Name:  "trajectory",
				Usage: "Common operations for working with trajectories.",
				Subcommands: []*cli.Command{
					{
						Name:  "plot",
						Usage: "Plot ballistics tables for custom rifle configurations and different cartridges",
						Flags: flagset.ExtractPrefix("em", plotTrajectory),
						Action: func(ctx *cli.Context) error {
							//todo: support loading this from a file

							var header []string
							var g, p float64
							var baseLength length.Length
							var baseMass mass.Mass
							var distance length.Length
							var drop length.Length

							// Pressure SI - N/(m^2)
							// Pressure Imperial - lb / (in ^ 2)

							switch plotTrajectory.Unit {
							case "imperial":
								header = []string{"yards", "time", "drop (in)", "velocity (ft/s)", "energy (ft-lbs)"}

								g = gravity.EarthImperial
								p = 0.0752

								baseLength = length.Inch
								baseMass = mass.Pound
								distance = length.Yard
								drop = length.Inch
							case "si":
								header = []string{"meters", "time", "drop (cm)", "velocity (ft/s)", "energy (ft-lbs)"}

								g = gravity.EarthSI
								p = 1.204

								baseLength = length.Meter
								baseMass = mass.Kilogram
								distance = length.Meter
								drop = length.Centimeter
							default:
								return fmt.Errorf("unsupported unit: %s", plotTrajectory.Unit)
							}

							out := csv.NewWriter(ctx.App.Writer)
							out.Comma = '\t'

							defer out.Flush()

							calc := Calculator{}

							// currently, pressure is PSI, eventually, we need to convert that to something SI
							diameter := plotTrajectory.BulletCaliber.As(baseLength)
							weight := plotTrajectory.BulletWeight.As(baseMass)
							barrelLength := plotTrajectory.BarrelLength.As(baseLength)

							slugs := plotTrajectory.BulletWeight.As(Slug)

							projectileAcceleration := calc.ProjectileAcceleration(plotTrajectory.CartridgePressure, diameter, weight)
							v0 := calc.MuzzleVelocity(projectileAcceleration, barrelLength)

							theta := 0.0
							if plotTrajectory.ZeroDistance > 0 {
								theta = math.Asin(g*plotTrajectory.ZeroDistance/math.Pow(v0, 2)) / 2
							}

							area := math.Pi * math.Pow(diameter/2, 2)
							plot := calc.Trajectory(g, v0, theta, func(v float64) float64 {
								// todo: this is no where near close
								// replace with actual drag function that can produce an accurate deceleration due to drag
								return 0.5 * p * plotTrajectory.DragCoefficient * area * math.Pow(v, 2)
							})

							_ = out.Write(header)
							for x := 0; x <= plotTrajectory.Range; x += plotTrajectory.Step {
								t, y, velocity := plot(x)
								energy := calc.ProjectileEnergy(velocity, slugs)

								_ = out.Write([]string{
									strconv.FormatInt(int64(x), 10),
									strconv.FormatFloat(t, 'f', 2, 64),
									strconv.FormatFloat(y*distance.As(drop), 'f', 2, 64),
									strconv.FormatFloat(velocity, 'f', 2, 64),
									strconv.FormatFloat(energy, 'f', 2, 64),
								})
							}

							return nil
						},
					},
				},
			},
		},
	}
)
