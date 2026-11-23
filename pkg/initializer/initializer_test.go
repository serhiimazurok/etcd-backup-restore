// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package initializer

import (
	"context"
	"errors"

	"github.com/gardener/etcd-backup-restore/pkg/member"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeMemberControl struct {
	member.Control
	noMemberReadyErr error
	scaledUpErr      error
	calls            []string
	noMemberReady    bool
	leaseHolder      bool
	memberInCluster  bool
	scaledUp         bool
}

func (f *fakeMemberControl) NoMemberReady(_ context.Context) (bool, error) {
	f.calls = append(f.calls, "NoMemberReady")
	return f.noMemberReady, f.noMemberReadyErr
}

func (f *fakeMemberControl) HasMemberLeaseHolder(_ context.Context, _ client.Client) bool {
	f.calls = append(f.calls, "HasMemberLeaseHolder")
	return f.leaseHolder
}

func (f *fakeMemberControl) WasMemberInCluster(_ context.Context, _ client.Client) bool {
	f.calls = append(f.calls, "WasMemberInCluster")
	return f.memberInCluster
}

func (f *fakeMemberControl) IsClusterScaledUp(_ context.Context) (bool, error) {
	f.calls = append(f.calls, "IsClusterScaledUp")
	return f.scaledUp, f.scaledUpErr
}

var _ = DescribeTable("checkMembership",
	func(m *fakeMemberControl, expectedHeartbeat, expectedScaleup bool, expectedCalls []string) {
		heartbeat, scaleup := checkMembership(context.Background(), logrus.New().WithField("suite", "initializer"), m, fake.NewClientBuilder().Build())
		Expect(heartbeat).To(Equal(expectedHeartbeat))
		Expect(scaleup).To(Equal(expectedScaleup))
		Expect(m.calls).To(Equal(expectedCalls))
	},
	Entry("no member ready and the lease names a holder: member was in the cluster, no probes",
		&fakeMemberControl{noMemberReady: true, leaseHolder: true, memberInCluster: true, scaledUp: true},
		true, false, []string{"NoMemberReady", "HasMemberLeaseHolder"}),
	Entry("no member ready and the lease names no holder: no scale-up, no probes",
		&fakeMemberControl{noMemberReady: true, memberInCluster: true, scaledUp: true},
		false, false, []string{"NoMemberReady", "HasMemberLeaseHolder"}),
	Entry("member readiness unknown: probes the cluster",
		&fakeMemberControl{noMemberReady: true, noMemberReadyErr: errors.New("pods is forbidden"), scaledUp: true},
		false, true, []string{"NoMemberReady", "WasMemberInCluster", "IsClusterScaledUp"}),
	Entry("a member ready and the member in the cluster: skips the scale-up check",
		&fakeMemberControl{memberInCluster: true, scaledUp: true},
		true, false, []string{"NoMemberReady", "WasMemberInCluster"}),
	Entry("a member ready and the member not in the cluster: detects the scale-up",
		&fakeMemberControl{scaledUp: true},
		false, true, []string{"NoMemberReady", "WasMemberInCluster", "IsClusterScaledUp"}),
	Entry("a member ready and the scale-up check failing: no scale-up",
		&fakeMemberControl{scaledUp: true, scaledUpErr: errors.New("etcd unavailable")},
		false, false, []string{"NoMemberReady", "WasMemberInCluster", "IsClusterScaledUp"}),
)
