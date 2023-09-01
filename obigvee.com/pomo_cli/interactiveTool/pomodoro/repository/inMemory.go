package repository

/**
* This module implements a data store for Pomodoro interval using Repository pattern.
* This helps to decouple the data store implementation from the business logic to bring
* flexibility to decision of how to store data, this allows modification to switch to a different 
* database entirely without affecting the business logic.
*/

import (
	"fmt"
	"sync"

	"obigvee.com/pomo_cli/interactiveTool/pomo/obigvee.com/pomo_cli/interactiveTool/pomodoro"
)

type inMemoryRepo struct {
	sync.RWMutex // mutexes prevents concurrent access to data
	intervals [] pomodoro.Interval
}

func NewInMemoryRepo() *inMemoryRepo {
	/**
	* NewInMemoryRepo - function instantiates a new inMemoryRepo type wih empty slice of type pomodoro.interval
	* Return : instance of slice of pomodoro.interval
	*/
	return &inMemoryRepo{
		intervals: []pomodoro.Interval{},
	}
}

// Implementation of all the methods of the Repository interface using inMemoryRepo type

func (r *inMemoryRepo) Create (i pomodoro.Interval) (int64, error){
	/**
	* Create - method takes instance of pomodoro.interval as input, save the values to the data store 
	* 
	* Return: ID of the saved entry
	*/
	
	r.Lock() // prevents concurrent access to the data store while making changes to it.
	defer r.Unlock()

	i.ID = int64(len(r.intervals)) + 1

	r.intervals = append(r.intervals, i)

	return i.ID, nil
}

func (r *inMemoryRepo)  Update(i pomodoro.Interval) error {
	/**
	* Update - method updates the values of an existing entry in the data store.
	*/
	
	r.Lock()
	defer r.Unlock()
	if i.ID == 0 {
		return fmt.Errorf("%w: %d", pomodoro.ErrInvalidID, i.ID)
	}
	
	r.intervals[i.ID-1] = i
	return nil
}

func (r *inMemoryRepo) ByID(id int64)(pomodoro.Interval, error) {
	r.RLock()
	defer r.RUnlock()
	i := pomodoro.Interval{}
	if id == 0 {
		return i, fmt.Errorf("%w: %d", pomodoro.ErrInvalidID, id)
	}
	
	i = r.intervals[id-1]
	return i, nil
}

func (r *inMemoryRepo) Breaks(n int) ([]pomodoro.Interval, error)  {
	r.RLocker()
	defer r.RUnlock()
	data := []pomodoro.Interval{}
	for k := len(r.intervals) - 1; k >= 0; k-- {
		if r.intervals[k].Category == pomodoro.CategoryPomodoro {
			continue
		}
		data = append(data, r.intervals[k])
		if len(data) == n{
			return data, nil
		}
	}
	
	return data, nil
}
