package main

import "time"

// TimeSlice is a of time.Time which implements the sort.Sort interface.
type TimeSlice []time.Time

// Append is a helper to append a time.Time to a *TimeSlice
func (ts *TimeSlice) Append(t time.Time) {
	*ts = append(*ts, t)
}

// Len returns the length of a TimeSlice for sorting
func (ts TimeSlice) Len() int { return len(ts) }

// Less returns true if i is newer than j for sorting. Sorting the TimeSlice
// will order it from newest to oldest.
func (ts TimeSlice) Less(i, j int) bool { return ts[i].After(ts[j]) }

// Swap swaps i and j in a TimeSlice for sorting.
func (ts TimeSlice) Swap(i, j int) { ts[i], ts[j] = ts[j], ts[i] }
