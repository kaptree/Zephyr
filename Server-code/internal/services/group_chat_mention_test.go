package services

import "testing"

// TestFilterMentionIDs @ 提及 ID 过滤：去重、保序、仅保留真实群成员
func TestFilterMentionIDs(t *testing.T) {
	members := []string{"u1", "u2", "u3"}

	cases := []struct {
		name     string
		mentions []string
		want     []string
	}{
		{"空提及", nil, nil},
		{"空数组", []string{}, nil},
		{"全部有效", []string{"u2", "u1"}, []string{"u2", "u1"}},
		{"去重保持首现顺序", []string{"u3", "u1", "u3", "u2", "u1"}, []string{"u3", "u1", "u2"}},
		{"过滤非成员", []string{"u1", "ghost", "u2"}, []string{"u1", "u2"}},
		{"过滤空字符串", []string{"", "u1", ""}, []string{"u1"}},
		{"全为非成员", []string{"ghost1", "ghost2"}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FilterMentionIDs(members, tc.mentions)
			if len(got) != len(tc.want) {
				t.Fatalf("FilterMentionIDs(%v) = %v, want %v", tc.mentions, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("FilterMentionIDs(%v) = %v, want %v", tc.mentions, got, tc.want)
				}
			}
		})
	}
}
