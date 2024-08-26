/*
 * Strava API v3
 *
 * The [Swagger Playground](https://developers.strava.com/playground) is the easiest way to familiarize yourself with the Strava API by submitting HTTP requests and observing the responses before you write any client code. It will show what a response will look like with different endpoints depending on the authorization scope you receive from your athletes. To use the Playground, go to https://www.strava.com/settings/api and change your “Authorization Callback Domain” to developers.strava.com. Please note, we only support Swagger 2.0. There is a known issue where you can only select one scope at a time. For more information, please check the section “client code” at https://developers.strava.com/docs.
 *
 * API version: 3.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package stravaapi

// SportType : An enumeration of the sport types an activity may have. Distinct from ActivityType in that it has new types (e.g. MountainBikeRide).
type SportType string

// List of SportType.
const (
	ALPINE_SKI_SportType                       SportType = "AlpineSki"
	BACKCOUNTRY_SKI_SportType                  SportType = "BackcountrySki"
	BADMINTON_SportType                        SportType = "Badminton"
	CANOEING_SportType                         SportType = "Canoeing"
	CROSSFIT_SportType                         SportType = "Crossfit"
	E_BIKE_RIDE_SportType                      SportType = "EBikeRide"
	ELLIPTICAL_SportType                       SportType = "Elliptical"
	E_MOUNTAIN_BIKE_RIDE_SportType             SportType = "EMountainBikeRide"
	GOLF_SportType                             SportType = "Golf"
	GRAVEL_RIDE_SportType                      SportType = "GravelRide"
	HANDCYCLE_SportType                        SportType = "Handcycle"
	HIGH_INTENSITY_INTERVAL_TRAINING_SportType SportType = "HighIntensityIntervalTraining"
	HIKE_SportType                             SportType = "Hike"
	ICE_SKATE_SportType                        SportType = "IceSkate"
	INLINE_SKATE_SportType                     SportType = "InlineSkate"
	KAYAKING_SportType                         SportType = "Kayaking"
	KITESURF_SportType                         SportType = "Kitesurf"
	MOUNTAIN_BIKE_RIDE_SportType               SportType = "MountainBikeRide"
	NORDIC_SKI_SportType                       SportType = "NordicSki"
	PICKLEBALL_SportType                       SportType = "Pickleball"
	PILATES_SportType                          SportType = "Pilates"
	RACQUETBALL_SportType                      SportType = "Racquetball"
	RIDE_SportType                             SportType = "Ride"
	ROCK_CLIMBING_SportType                    SportType = "RockClimbing"
	ROLLER_SKI_SportType                       SportType = "RollerSki"
	ROWING_SportType                           SportType = "Rowing"
	RUN_SportType                              SportType = "Run"
	SAIL_SportType                             SportType = "Sail"
	SKATEBOARD_SportType                       SportType = "Skateboard"
	SNOWBOARD_SportType                        SportType = "Snowboard"
	SNOWSHOE_SportType                         SportType = "Snowshoe"
	SOCCER_SportType                           SportType = "Soccer"
	SQUASH_SportType                           SportType = "Squash"
	STAIR_STEPPER_SportType                    SportType = "StairStepper"
	STAND_UP_PADDLING_SportType                SportType = "StandUpPaddling"
	SURFING_SportType                          SportType = "Surfing"
	SWIM_SportType                             SportType = "Swim"
	TABLE_TENNIS_SportType                     SportType = "TableTennis"
	TENNIS_SportType                           SportType = "Tennis"
	TRAIL_RUN_SportType                        SportType = "TrailRun"
	VELOMOBILE_SportType                       SportType = "Velomobile"
	VIRTUAL_RIDE_SportType                     SportType = "VirtualRide"
	VIRTUAL_ROW_SportType                      SportType = "VirtualRow"
	VIRTUAL_RUN_SportType                      SportType = "VirtualRun"
	WALK_SportType                             SportType = "Walk"
	WEIGHT_TRAINING_SportType                  SportType = "WeightTraining"
	WHEELCHAIR_SportType                       SportType = "Wheelchair"
	WINDSURF_SportType                         SportType = "Windsurf"
	WORKOUT_SportType                          SportType = "Workout"
	YOGA_SportType                             SportType = "Yoga"
)
