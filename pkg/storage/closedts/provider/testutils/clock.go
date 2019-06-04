// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License included
// in the file licenses/BSL.txt and at www.mariadb.com/bsl11.
//
// Change Date: 2022-10-01
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt and at
// https://www.apache.org/licenses/LICENSE-2.0

package testutils

import (
	"errors"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/closedts/ctpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

// A TestClock provides a blocking LiveClockFn that can be triggered
// at will.
type TestClock struct {
	stopper *stop.Stopper
	ch      chan tick
}

// NewTestClock sets up a test clock that returns errors once the
// Stopper starts quiescing.
func NewTestClock(stopper *stop.Stopper) *TestClock {
	t := &TestClock{
		stopper: stopper,
		ch:      make(chan tick),
	}
	return t
}

type tick struct {
	liveNow   hlc.Timestamp
	liveEpoch ctpb.Epoch
	err       error
}

// Tick is called by tests to manually emit a single clock tick. The tick
// will only returned to a single caller of LiveNow().
func (c *TestClock) Tick(liveNow hlc.Timestamp, liveEpoch ctpb.Epoch, err error) {
	c.ch <- tick{liveNow, liveEpoch, err}
}

// LiveNow implements closedts.LiveClockFn.
func (c *TestClock) LiveNow(roachpb.NodeID) (liveNow hlc.Timestamp, liveEpoch ctpb.Epoch, _ error) {
	select {
	case r := <-c.ch:
		return r.liveNow, r.liveEpoch, r.err
	case <-c.stopper.ShouldQuiesce():
		return hlc.Timestamp{}, 0, errors.New("quiescing")
	}
}
