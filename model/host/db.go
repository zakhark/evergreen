package host

import (
	"fmt"
	"time"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/db"
	"github.com/evergreen-ci/evergreen/db/bsonutil"
	"github.com/evergreen-ci/evergreen/model/distro"
	"github.com/evergreen-ci/evergreen/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	// Collection is the name of the MongoDB collection that stores hosts.
	Collection = "hosts"
)

var noRunningTask = []bson.D{
	{{RunningTaskKey, ""}},
	{{RunningTaskKey, bson.D{{"$exists", false}}}},
}

var (
	IdKey                    = bsonutil.MustHaveTag(Host{}, "Id")
	DNSKey                   = bsonutil.MustHaveTag(Host{}, "Host")
	SecretKey                = bsonutil.MustHaveTag(Host{}, "Secret")
	UserKey                  = bsonutil.MustHaveTag(Host{}, "User")
	TagKey                   = bsonutil.MustHaveTag(Host{}, "Tag")
	DistroKey                = bsonutil.MustHaveTag(Host{}, "Distro")
	ProviderKey              = bsonutil.MustHaveTag(Host{}, "Provider")
	ProvisionedKey           = bsonutil.MustHaveTag(Host{}, "Provisioned")
	RunningTaskKey           = bsonutil.MustHaveTag(Host{}, "RunningTask")
	PidKey                   = bsonutil.MustHaveTag(Host{}, "Pid")
	TaskDispatchTimeKey      = bsonutil.MustHaveTag(Host{}, "TaskDispatchTime")
	CreateTimeKey            = bsonutil.MustHaveTag(Host{}, "CreationTime")
	ExpirationTimeKey        = bsonutil.MustHaveTag(Host{}, "ExpirationTime")
	TerminationTimeKey       = bsonutil.MustHaveTag(Host{}, "TerminationTime")
	LTCTimeKey               = bsonutil.MustHaveTag(Host{}, "LastTaskCompletedTime")
	LTCKey                   = bsonutil.MustHaveTag(Host{}, "LastTaskCompleted")
	StatusKey                = bsonutil.MustHaveTag(Host{}, "Status")
	AgentRevisionKey         = bsonutil.MustHaveTag(Host{}, "AgentRevision")
	StartedByKey             = bsonutil.MustHaveTag(Host{}, "StartedBy")
	InstanceTypeKey          = bsonutil.MustHaveTag(Host{}, "InstanceType")
	NotificationsKey         = bsonutil.MustHaveTag(Host{}, "Notifications")
	UserDataKey              = bsonutil.MustHaveTag(Host{}, "UserData")
	LastReachabilityCheckKey = bsonutil.MustHaveTag(Host{}, "LastReachabilityCheck")
	LastCommunicationTimeKey = bsonutil.MustHaveTag(Host{}, "LastCommunicationTime")
	UnreachableSinceKey      = bsonutil.MustHaveTag(Host{}, "UnreachableSince")
)

// === Queries ===

// All is a query that returns all hosts
var All = db.Query(nil)

// ByUserWithRunningStatus produces a query that returns all
// running hosts for the given user id.
func ByUserWithRunningStatus(user string) db.Q {
	return db.Query(
		bson.M{
			StartedByKey: user,
			StatusKey:    bson.M{"$ne": evergreen.HostTerminated},
		})
}

// IsRunning is a query that returns all hosts that are running
// (i.e. status != terminated).
var IsRunning = db.Query(bson.M{StatusKey: bson.M{"$ne": evergreen.HostTerminated}})

// IsLive is a query that returns all working hosts started by Evergreen
var IsLive = db.Query(
	bson.M{
		StartedByKey: evergreen.User,
		StatusKey:    bson.M{"$in": evergreen.UphostStatus},
	},
)

// ByUserWithUnterminatedStatus produces a query that returns all running hosts
// for the given user id.
func ByUserWithUnterminatedStatus(user string) db.Q {
	return db.Query(
		bson.M{
			StartedByKey: user,
			StatusKey:    bson.M{"$ne": evergreen.HostTerminated},
		},
	)
}

// IsAvailableAndFree is a query that returns all running
// Evergreen hosts without an assigned task.
var IsAvailableAndFree = db.Query(
	bson.M{
		"$or":        noRunningTask,
		StatusKey:    evergreen.HostRunning,
		StartedByKey: evergreen.User,
	},
).Sort([]string{"-" + LTCTimeKey})

// ByAvailableForDistro returns all running Evergreen hosts with
// no running task of a certain distro Id.
func ByAvailableForDistro(d string) db.Q {
	distroIdKey := fmt.Sprintf("%v.%v", DistroKey, distro.IdKey)
	return db.Query(bson.M{
		distroIdKey:  d,
		"$or":        noRunningTask,
		StatusKey:    evergreen.HostRunning,
		StartedByKey: evergreen.User,
	}).Sort([]string{"-" + LTCTimeKey})
}

// IsFree is a query that returns all running
// Evergreen hosts without an assigned task.
var IsFree = db.Query(
	bson.M{
		"$or":        noRunningTask,
		StartedByKey: evergreen.User,
		StatusKey:    evergreen.HostRunning,
	},
)

// ByUnprovisionedSince produces a query that returns all hosts
// Evergreen never finished setting up that were created before
// the given time.
func ByUnprovisionedSince(threshold time.Time) db.Q {
	return db.Query(bson.M{
		ProvisionedKey: false,
		CreateTimeKey:  bson.M{"$lte": threshold},
		StatusKey:      bson.M{"$ne": evergreen.HostTerminated},
		StartedByKey:   evergreen.User,
	})
}

// IsUninitialized is a query that returns all uninitialized Evergreen hosts.
var IsUninitialized = db.Query(
	bson.M{StatusKey: evergreen.HostUninitialized},
)

// ByUnproductiveSince produces a query that returns all hosts that
// are not doing work and were created before the given time.
func ByUnproductiveSince(threshold time.Time) db.Q {
	return db.Query(bson.M{
		"$or":         noRunningTask,
		LTCKey:        "",
		CreateTimeKey: bson.M{"$lte": threshold},
		StatusKey:     bson.M{"$ne": evergreen.HostTerminated},
		StartedByKey:  evergreen.User,
	})
}

// IsRunningAndSpawned is a query that returns all running hosts
// spawned by an Evergreen user.
var IsRunningAndSpawned = db.Query(
	bson.M{
		StartedByKey: bson.M{"$ne": evergreen.User},
		StatusKey:    bson.M{"$ne": evergreen.HostTerminated},
	},
)

// IsRunningTask is a query that returns all running hosts with a running task
var IsRunningTask = db.Query(
	bson.M{
		RunningTaskKey: bson.M{"$exists": true},
	},
)

// IsDecommissioned is a query that returns all hosts without a
// running task that are marked for decommissioning.
var IsDecommissioned = db.Query(
	bson.M{
		"$or":     noRunningTask,
		StatusKey: evergreen.HostDecommissioned},
)

// ByDistroId produces a query that returns all working hosts (not terminated and
// not quarantined) of the given distro.
func ByDistroId(distroId string) db.Q {
	dId := fmt.Sprintf("%v.%v", DistroKey, distro.IdKey)
	return db.Query(bson.M{
		dId:          distroId,
		StartedByKey: evergreen.User,
		StatusKey:    bson.M{"$in": evergreen.UphostStatus},
	})
}

// ById produces a query that returns a host with the given id.
func ById(id string) db.Q {
	return db.Query(bson.D{{IdKey, id}})
}

// ByIds produces a query that returns all hosts in the given list of ids.
func ByIds(ids []string) db.Q {
	return db.Query(bson.D{
		{IdKey, bson.D{{"$in", ids}}},
	})
}

// ByRunningTaskId returns a host running the task with the given id.
func ByRunningTaskId(taskId string) db.Q {
	return db.Query(bson.D{{RunningTaskKey, taskId}})
}

// ByDynamicWithinTime is a query that returns all dynamic hosts running between a certain time and another time.
func ByDynamicWithinTime(startTime, endTime time.Time) db.Q {
	return db.Query(
		bson.M{
			"$or": []bson.M{
				bson.M{
					CreateTimeKey:      bson.M{"$lt": endTime},
					TerminationTimeKey: bson.M{"$gt": startTime},
					ProviderKey:        bson.M{"$ne": evergreen.HostTypeStatic},
				},
				bson.M{
					CreateTimeKey:      bson.M{"$lt": endTime},
					TerminationTimeKey: util.ZeroTime,
					StatusKey:          evergreen.HostRunning,
					ProviderKey:        bson.M{"$ne": evergreen.HostTypeStatic},
				},
			},
		})
}

var AllStatic = db.Query(
	bson.M{
		ProviderKey: evergreen.HostTypeStatic,
	})

// IsIdle is a query that returns all running Evergreen hosts with no task.
var IsIdle = db.Query(
	bson.M{
		"$or":        noRunningTask,
		StatusKey:    evergreen.HostRunning,
		StartedByKey: evergreen.User,
	},
)

// IsActive is a query that returns all Evergreen hosts that are working or
// capable of being assigned work to do.
var IsActive = db.Query(
	bson.M{
		StartedByKey: evergreen.User,
		StatusKey: bson.M{
			"$nin": []string{
				evergreen.HostTerminated, evergreen.HostDecommissioned, evergreen.HostInitializing,
			},
		},
	},
)

// ByNotMonitoredSince produces a query that returns all hosts whose
// last reachability check was before the specified threshold,
// filtering out user-spawned hosts and hosts currently running tasks.
func ByNotMonitoredSince(threshold time.Time) db.Q {
	return db.Query(bson.M{
		"$and": []bson.M{
			{"$or": noRunningTask},
			{StatusKey: bson.M{
				"$in": []string{evergreen.HostRunning, evergreen.HostUnreachable},
			}},
			{StartedByKey: evergreen.User},
			{"$or": []bson.M{
				{LastReachabilityCheckKey: bson.M{"$lte": threshold}},
				{LastReachabilityCheckKey: bson.M{"$exists": false}},
			}},
		},
	})
}

// ByExpiringBetween produces a query that returns  any user-spawned hosts
// that will expire between the specified times.
func ByExpiringBetween(lowerBound time.Time, upperBound time.Time) db.Q {
	return db.Query(bson.M{
		StartedByKey: bson.M{"$ne": evergreen.User},
		StatusKey: bson.M{
			"$nin": []string{evergreen.HostTerminated, evergreen.HostQuarantined},
		},
		ExpirationTimeKey: bson.M{"$gte": lowerBound, "$lte": upperBound},
	})
}

// ByUnreachableBefore produces a query that returns a list of all
// hosts that are still unreachable, and have been in that state since before the
// given time threshold.
func ByUnreachableBefore(threshold time.Time) db.Q {
	return db.Query(bson.M{
		StatusKey:           evergreen.HostUnreachable,
		UnreachableSinceKey: bson.M{"$gt": time.Unix(0, 0), "$lt": threshold},
	})
}

// ByExpiredSicne produces a query that returns any user-spawned hosts
// that will expired after the given time.
func ByExpiredSince(time time.Time) db.Q {
	return db.Query(bson.M{
		StartedByKey: bson.M{"$ne": evergreen.User},
		StatusKey: bson.M{
			"$nin": []string{evergreen.HostTerminated, evergreen.HostQuarantined},
		},
		ExpirationTimeKey: bson.M{"$lte": time},
	})
}

// IsProvisioningFailure is a query that returns all hosts that
// failed to provision.
var IsProvisioningFailure = db.Query(bson.D{{StatusKey, evergreen.HostProvisionFailed}})

// === DB Logic ===

// FindOne gets one Host for the given query.
func FindOne(query db.Q) (*Host, error) {
	host := &Host{}
	err := db.FindOneQ(Collection, query, host)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	return host, err
}

// Find gets all Hosts for the given query.
func Find(query db.Q) ([]Host, error) {
	hosts := []Host{}
	err := db.FindAllQ(Collection, query, &hosts)
	return hosts, err
}

// Count returns the number of hosts that satisfy the given query.
func Count(query db.Q) (int, error) {
	return db.CountQ(Collection, query)
}

// UpdateOne updates one host.
func UpdateOne(query interface{}, update interface{}) error {
	return db.Update(
		Collection,
		query,
		update,
	)
}

// UpdateAll updates all hosts.
func UpdateAll(query interface{}, update interface{}) error {
	_, err := db.UpdateAll(
		Collection,
		query,
		update,
	)
	return err
}

// UpsertOne upserts a host.
func UpsertOne(query interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	return db.Upsert(
		Collection,
		query,
		update,
	)
}
