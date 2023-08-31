package pomodoro

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Category constants
const (
	CategoryPomodoro = "Pomodoro"
	CategoryShortBreak = "ShortBreak"
	CategoryLongBreak = "LongBreak"
)

// State constants
const (
	StateNotStarted = iota
	StateRunning
	StatePaused
	StateDone
	StateCancelled
)

// interval struct

type Interval struct{
	ID int64
	StartTime time.Time
	PlannedDuration time.Duration
	ActualDuration time.Duration
	Category string
	State int
}

// define Repo interface
type Repository interface{
	Create(i Interval)(int64, error) // create/saves a new interval
	Update(i Interval)(error) // update details about an interval
	ByID(id int64)(Interval, error) // retrieve an interval by ID
	Last() (Interval, error) // find the last interval and retrieve it
	Breaks(n int) ([]Interval, error) // retrieve a given number of interval
}


/**
 * define error flags values ro rep particular errors that it may return
 */ 
var (
	ErrNoIntervals = errors.New("No Intervals")
	ErrIntervalNotRunning = errors.New("Interval not running")
	ErrIntervalCompleted = errors.New("Interval is completed or is cancelled")
	ErrInvalidState = errors.New("Invalid State")
	ErrInvalidID = errors.New("the ID is not valid, try another one")
)

type IntervalConfig struct{
	/**
	* IntervalConfig rep the config required to instantiate
	* n interval
	*/
	repo Repository
	PomodoroDuration time.Duration
	ShortBreakDuration time.Duration
	LongBreakDuration time.Duration
}


// instantiate new IntervalConfig
func NewConfig(repo Repository, pomodoro, shortBreak, longBreak time.Duration) *IntervalConfig{
	c:= &IntervalConfig{
		repo: repo,
		PomodoroDuration: 25 * time.Minute,
		ShortBreakDuration:  5 * time.Minute,
		LongBreakDuration: 15 * time.Minute,
	}
	
	if pomodoro > 0{
		c.PomodoroDuration = pomodoro
	}

	if shortBreak > 0{
		c.ShortBreakDuration = shortBreak
	}
	if longBreak > 0{
		c.LongBreakDuration = longBreak
	}
	return c
}

func nextCategory(r Repository) (string, error) {
	li, err := r.Last()
	if err != nil && err == ErrNoIntervals{
		return CategoryPomodoro, nil
	}
	if err != nil{
		return "", err
	}
	if li.Category == CategoryLongBreak || li.Category == CategoryShortBreak{
		return CategoryPomodoro, nil
	}
	lastBreaks, err := r.Breaks(3)
	if err != nil{
		return "", err
	}
	if len(lastBreaks) < 3{
		return CategoryLongBreak, nil
	}

	for _, i := range lastBreaks{
		if i.Category == CategoryLongBreak{
			return CategoryShortBreak, nil
		}
	}
	
	return CategoryLongBreak, nil
}

// Callback function accepts an instance of type interval as input return nothing
type Callback func(Interval)

func tick(ctx context.Context, id int64, config *IntervalConfig,
		start, periodic, end Callback) error {
			/**
			* tick - function  controls the timer for each interval's execution.
			* @ctx: instance of context.Context, it indicates a cancellation
			* @id: id of interval to control
			* @config: instance of the configuration IntervalConfig
			* @start: Callback function
			* @periodic: Callback function
			* @end: Callback function
			* Return : error
			*/

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		
		i, err := config.repo.ByID(id)
		if err != nil{
			return err
		}

		expire := time.After(i.PlannedDuration - i.ActualDuration)
		start(i)

		for{
			select {
			case <-ticker.C:
				i, err := config.repo.ByID(id)
				if err != nil{
					return err
				}
				if i.State == StatePaused{
					return nil
				}
				
				i.ActualDuration += time.Second
				if err := config.repo.Update(i); err != nil{
					return err
				}
				periodic(i)
			case <-expire:
				i, err := config.repo.ByID(id)
				if err != nil {
					return err
				}
				i.State = StateDone
				end(i)
				return config.repo.Update(i)
			case <-ctx.Done():
				i, err := config.repo.ByID(id)
				if err != nil{
					return err
				}
				i.State = StateCancelled
				return config.repo.Update(i)
			}
		}
}

func newInterval(config *IntervalConfig) (Interval, error) {
/**
* newInterval - function takes an instance of the config intervalConfig 
* @config: an instance of the intervalConfig
* 
* Returns: a interval instance with appropriate category and values
*/
	i := Interval{}
	category, err := nextCategory(config.repo)
	if err != nil {
		return i, err
	}

	i.Category = category
	
	switch category {
	case CategoryPomodoro:
		i.PlannedDuration = config.PomodoroDuration
	case CategoryShortBreak:
		i.PlannedDuration = config.ShortBreakDuration
	case CategoryLongBreak:
		i.PlannedDuration = config.LongBreakDuration
	}

	if i.ID, err = config.repo.Create(i); err != nil{
		return i, err
	}
	
	return i, nil
}

func GetInterVal(config *IntervalConfig) (Interval, error)  {
	/**
	* GetInterval - attempts to retrieve the last interval from the repository 
	* @config: instance of IntervalConfig
	* 
	* Return: Interval instance if it's active or error when there's an issue accessing the repository
			  if the last interval is inactive or unavailable, it returns a new interval using the
			  previously defined function newInterval()
	*/

	i := Interval{}
	var err error
	
	i, err = config.repo.Last()

	if err != nil && err != ErrNoIntervals {
		return i, err
	}
	
	if err == nil && i.State != StateCancelled && i.State != StateDone {
		return i, nil
	}
	
	return newInterval(config)
}

func (i Interval) Start(ctx context.Context, config *IntervalConfig,
	start, periodic, end Callback) error {
	/**
	* Start() - method is used by callers to start the interval timer.
				It, checks the state of the current interval setting the appropriate options
				and then calls the tick function to the time interval. The function
				takes the same input params as the tick() function, with  callbacks to pass
				to tick() when required.
	* @ctx: instance of context.Context
	* @config:instance of IntervalConfig
	* @ start, @periodic @ end : Callback function
	* Return: error
	*/
	switch i.State {
	case StateRunning:
		return nil
	case StateNotStarted:
		i.StartTime = time.Now()
		fallthrough
	case StatePaused:
		i.State = StateRunning
		if err := config.repo.Update(i); err != nil{
			return err
		}
		return tick(ctx, i.ID, config, start, periodic, end)
	case StateCancelled, StateDone:
		return fmt.Errorf("%w: Cannot start", ErrIntervalCompleted)
	default:
		return fmt.Errorf("%w: %d", ErrInvalidState, i.State)
	}
}

func (i Interval) Pause(config *IntervalConfig) error  {
	/**
	* Pause() - method allows callers to pause a running interval.
			it verifies whether the instance of interval is running and pauses it by setting
			the state to StatePaused
	* @config: instance of IntervalConfig
	* Returns: error
	*/
	if i.State != StateRunning {
		return ErrIntervalNotRunning
	}

	i.State = StatePaused

	return config.repo.Update(i)
}
