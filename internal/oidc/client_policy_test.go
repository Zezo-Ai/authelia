package oidc_test

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewClientAuthorizationPolicy(t *testing.T) {
	lanip, lan, err := net.ParseCIDR("192.168.2.1/24")

	require.NoError(t, err)

	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectPolicy
		expected oidc.ClientAuthorizationPolicy
		extra    func(t *testing.T, actual oidc.ClientAuthorizationPolicy)
	}{
		{
			"ShouldHandleStandardExample",
			schema.IdentityProvidersOpenIDConnectPolicy{
				DefaultPolicy: "two_factor",
				Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
					{
						Policy: "one_factor",
						Subjects: [][]string{
							{"user:john"},
							{"user:bob"},
						},
					},
					{
						Policy:   "one_factor",
						Networks: []*net.IPNet{lan},
					},
				},
			},
			oidc.ClientAuthorizationPolicy{
				Name:          "test",
				DefaultPolicy: authorization.TwoFactor,
				Rules: []oidc.ClientAuthorizationPolicyRule{
					{
						Subjects: []authorization.AccessControlSubjects{
							{
								Subjects: []authorization.SubjectMatcher{authorization.AccessControlUser{Name: "john"}},
							},
							{
								Subjects: []authorization.SubjectMatcher{authorization.AccessControlUser{Name: "bob"}},
							},
						},
						Policy: authorization.OneFactor,
					},
					{
						Networks: []*net.IPNet{lan},
						Policy:   authorization.OneFactor,
					},
				},
			},
			func(t *testing.T, actual oidc.ClientAuthorizationPolicy) {
				assert.Equal(t, authorization.TwoFactor, actual.GetRequiredLevel(authorization.Subject{}))
				assert.Equal(t, authorization.OneFactor, actual.GetRequiredLevel(authorization.Subject{Username: "john"}))
				assert.Equal(t, authorization.OneFactor, actual.GetRequiredLevel(authorization.Subject{Username: "bob"}))
				assert.Equal(t, authorization.OneFactor, actual.GetRequiredLevel(authorization.Subject{Username: "fred", IP: lanip}))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := oidc.NewClientAuthorizationPolicy("test", tc.have)

			assert.Equal(t, tc.expected, actual)

			if tc.extra != nil {
				tc.extra(t, actual)
			}
		})
	}
}

func TestNewClientConsentPolicy(t *testing.T) {
	val := func(duration time.Duration) *time.Duration {
		return &duration
	}

	testCases := []struct {
		name     string
		mode     string
		duration *time.Duration
		expected oidc.ClientConsentPolicy
		extra    func(t *testing.T, actual oidc.ClientConsentPolicy)
	}{
		{
			"ShouldParsePolicyExplicit",
			"explicit",
			nil,
			oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModeExplicit},
			func(t *testing.T, actual oidc.ClientConsentPolicy) {
				assert.Equal(t, "explicit", actual.String())
			},
		},
		{
			"ShouldParsePolicyImplicit",
			"implicit",
			nil,
			oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModeImplicit},
			func(t *testing.T, actual oidc.ClientConsentPolicy) {
				assert.Equal(t, "implicit", actual.String())
			},
		},
		{
			"ShouldParsePolicyPreConfigured",
			"pre-configured",
			val(time.Hour * 20),
			oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModePreConfigured, Duration: time.Hour * 20},
			func(t *testing.T, actual oidc.ClientConsentPolicy) {
				assert.Equal(t, "pre-configured", actual.String())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := oidc.NewClientConsentPolicy(tc.mode, tc.duration)
			assert.Equal(t, tc.expected, actual)

			if tc.extra != nil {
				tc.extra(t, actual)
			}
		})
	}

	assert.Equal(t, "", oidc.ClientConsentMode(-1).String())
}

func TestNewClientRequestedAudienceMode(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected oidc.ClientRequestedAudienceMode
	}{
		{
			"ShouldParsePolicyExplicit",
			"explicit",
			oidc.ClientRequestedAudienceModeExplicit,
		},
		{
			"ShouldParsePolicyImplicit",
			"implicit",
			oidc.ClientRequestedAudienceModeImplicit,
		},
		{
			"ShouldParsePolicyImplicitByDefault",
			"",
			oidc.ClientRequestedAudienceModeImplicit,
		},
		{
			"ShouldParsePolicyImplicitByDefaultBadName",
			"bad",
			oidc.ClientRequestedAudienceModeImplicit,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.NewClientRequestedAudienceMode(tc.have))
		})
	}

	assert.Equal(t, "", oidc.ClientRequestedAudienceMode(-1).String())
}

func TestMatchesSubjects(t *testing.T) {
	lanip, lan, err := net.ParseCIDR("192.168.2.1/24")

	require.NoError(t, err)

	lanip2, _, err := net.ParseCIDR("192.168.3.1/24")

	require.NoError(t, err)

	subjectAnonymous := authorization.Subject{}
	subjectNoGroup := authorization.Subject{Username: "user-no-group"}
	subjectGroupA := authorization.Subject{Username: "user-group-a", Groups: []string{"a"}}
	subjectGroupB := authorization.Subject{Username: "user-group-b", Groups: []string{"b"}}
	subjectGroupAB := authorization.Subject{Username: "user-group-ab", Groups: []string{"a", "b"}}
	subjectNoGroupNet := authorization.Subject{Username: "user-no-group-net", IP: lanip}
	subjectGroupANet := authorization.Subject{Username: "user-group-a-net", Groups: []string{"a"}, IP: lanip}
	subjectGroupBNet := authorization.Subject{Username: "user-group-b-net", Groups: []string{"b"}, IP: lanip}
	subjectGroupABNet := authorization.Subject{Username: "user-group-ab-net", Groups: []string{"a", "b"}, IP: lanip}
	subjectGroupABNet2 := authorization.Subject{Username: "user-group-ab-net2", Groups: []string{"a", "b"}, IP: lanip2}
	allSubjects := []authorization.Subject{
		subjectAnonymous,
		subjectNoGroup,
		subjectGroupA,
		subjectGroupB,
		subjectGroupAB,
		subjectNoGroupNet,
		subjectGroupANet,
		subjectGroupBNet,
		subjectGroupABNet,
		subjectGroupABNet2,
	}

	testCases := []struct {
		name     string
		rule     oidc.ClientAuthorizationPolicyRule
		subjects []authorization.Subject
	}{
		{
			"ShouldMatchEmptyRule",
			oidc.ClientAuthorizationPolicyRule{
				Policy: authorization.OneFactor,
			},
			allSubjects,
		},
		{
			"ShouldMatchRuleWithGroupA",
			oidc.ClientAuthorizationPolicyRule{
				Subjects: []authorization.AccessControlSubjects{
					{
						Subjects: []authorization.SubjectMatcher{authorization.AccessControlGroup{Name: "a"}},
					},
				},
				Policy: authorization.OneFactor,
			},
			[]authorization.Subject{subjectGroupA, subjectGroupAB, subjectGroupANet, subjectGroupABNet, subjectGroupABNet2},
		},
		{
			"ShouldMatchRuleWithGroupsAB",
			oidc.ClientAuthorizationPolicyRule{
				Subjects: []authorization.AccessControlSubjects{
					{
						Subjects: []authorization.SubjectMatcher{authorization.AccessControlGroup{Name: "a"}},
					},
					{
						Subjects: []authorization.SubjectMatcher{authorization.AccessControlGroup{Name: "b"}},
					},
				},
				Policy: authorization.OneFactor,
			},
			[]authorization.Subject{subjectGroupA, subjectGroupB, subjectGroupAB, subjectGroupANet, subjectGroupBNet, subjectGroupABNet, subjectGroupABNet2},
		},
		{
			"ShouldMatchRuleNet",
			oidc.ClientAuthorizationPolicyRule{
				Networks: []*net.IPNet{lan},
				Policy:   authorization.OneFactor,
			},
			[]authorization.Subject{subjectNoGroupNet, subjectGroupANet, subjectGroupBNet, subjectGroupABNet},
		},
		{
			"ShouldMatchRuleGroupANet",
			oidc.ClientAuthorizationPolicyRule{
				Subjects: []authorization.AccessControlSubjects{
					{
						Subjects: []authorization.SubjectMatcher{authorization.AccessControlGroup{Name: "a"}},
					},
				},
				Networks: []*net.IPNet{lan},
				Policy:   authorization.OneFactor,
			},
			[]authorization.Subject{subjectGroupANet, subjectGroupABNet},
		},
		{
			"ShouldMatchRuleGroupsABNet",
			oidc.ClientAuthorizationPolicyRule{
				Subjects: []authorization.AccessControlSubjects{
					{
						Subjects: []authorization.SubjectMatcher{authorization.AccessControlGroup{Name: "a"}},
					},
					{
						Subjects: []authorization.SubjectMatcher{authorization.AccessControlGroup{Name: "b"}},
					},
				},
				Networks: []*net.IPNet{lan},
				Policy:   authorization.OneFactor,
			},
			[]authorization.Subject{subjectGroupANet, subjectGroupBNet, subjectGroupABNet},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matchingSubjects := []authorization.Subject{}

			for _, subject := range allSubjects {
				if tc.rule.MatchesSubjects(subject) {
					matchingSubjects = append(matchingSubjects, subject)
				}
			}

			assert.Equal(t, tc.subjects, matchingSubjects)
		})
	}
}
