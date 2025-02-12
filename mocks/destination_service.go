// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"sync"

	"github.com/libsv/payd"
)

// Ensure, that DestinationsServiceMock does implement payd.DestinationsService.
// If this is not the case, regenerate this file with moq.
var _ payd.DestinationsService = &DestinationsServiceMock{}

// DestinationsServiceMock is a mock implementation of payd.DestinationsService.
//
// 	func TestSomethingThatUsesDestinationsService(t *testing.T) {
//
// 		// make and configure a mocked payd.DestinationsService
// 		mockedDestinationsService := &DestinationsServiceMock{
// 			DestinationsFunc: func(ctx context.Context, args payd.DestinationsArgs) (*payd.Destination, error) {
// 				panic("mock out the Destinations method")
// 			},
// 			DestinationsCreateFunc: func(ctx context.Context, req payd.DestinationsCreate) (*payd.Destination, error) {
// 				panic("mock out the DestinationsCreate method")
// 			},
// 		}
//
// 		// use mockedDestinationsService in code that requires payd.DestinationsService
// 		// and then make assertions.
//
// 	}
type DestinationsServiceMock struct {
	// DestinationsFunc mocks the Destinations method.
	DestinationsFunc func(ctx context.Context, args payd.DestinationsArgs) (*payd.Destination, error)

	// DestinationsCreateFunc mocks the DestinationsCreate method.
	DestinationsCreateFunc func(ctx context.Context, req payd.DestinationsCreate) (*payd.Destination, error)

	// calls tracks calls to the methods.
	calls struct {
		// Destinations holds details about calls to the Destinations method.
		Destinations []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args payd.DestinationsArgs
		}
		// DestinationsCreate holds details about calls to the DestinationsCreate method.
		DestinationsCreate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Req is the req argument value.
			Req payd.DestinationsCreate
		}
	}
	lockDestinations       sync.RWMutex
	lockDestinationsCreate sync.RWMutex
}

// Destinations calls DestinationsFunc.
func (mock *DestinationsServiceMock) Destinations(ctx context.Context, args payd.DestinationsArgs) (*payd.Destination, error) {
	if mock.DestinationsFunc == nil {
		panic("DestinationsServiceMock.DestinationsFunc: method is nil but DestinationsService.Destinations was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args payd.DestinationsArgs
	}{
		Ctx:  ctx,
		Args: args,
	}
	mock.lockDestinations.Lock()
	mock.calls.Destinations = append(mock.calls.Destinations, callInfo)
	mock.lockDestinations.Unlock()
	return mock.DestinationsFunc(ctx, args)
}

// DestinationsCalls gets all the calls that were made to Destinations.
// Check the length with:
//     len(mockedDestinationsService.DestinationsCalls())
func (mock *DestinationsServiceMock) DestinationsCalls() []struct {
	Ctx  context.Context
	Args payd.DestinationsArgs
} {
	var calls []struct {
		Ctx  context.Context
		Args payd.DestinationsArgs
	}
	mock.lockDestinations.RLock()
	calls = mock.calls.Destinations
	mock.lockDestinations.RUnlock()
	return calls
}

// DestinationsCreate calls DestinationsCreateFunc.
func (mock *DestinationsServiceMock) DestinationsCreate(ctx context.Context, req payd.DestinationsCreate) (*payd.Destination, error) {
	if mock.DestinationsCreateFunc == nil {
		panic("DestinationsServiceMock.DestinationsCreateFunc: method is nil but DestinationsService.DestinationsCreate was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Req payd.DestinationsCreate
	}{
		Ctx: ctx,
		Req: req,
	}
	mock.lockDestinationsCreate.Lock()
	mock.calls.DestinationsCreate = append(mock.calls.DestinationsCreate, callInfo)
	mock.lockDestinationsCreate.Unlock()
	return mock.DestinationsCreateFunc(ctx, req)
}

// DestinationsCreateCalls gets all the calls that were made to DestinationsCreate.
// Check the length with:
//     len(mockedDestinationsService.DestinationsCreateCalls())
func (mock *DestinationsServiceMock) DestinationsCreateCalls() []struct {
	Ctx context.Context
	Req payd.DestinationsCreate
} {
	var calls []struct {
		Ctx context.Context
		Req payd.DestinationsCreate
	}
	mock.lockDestinationsCreate.RLock()
	calls = mock.calls.DestinationsCreate
	mock.lockDestinationsCreate.RUnlock()
	return calls
}
