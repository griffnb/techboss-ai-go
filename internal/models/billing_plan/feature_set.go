package billing_plan

const (
	FEATURE_DISABLED int64 = 0
	FEATURE_ENABLED  int64 = 1
)

type FeatureSet struct{}

type MergeableFeatureSet struct{}

func (this *FeatureSet) Merge(_ *MergeableFeatureSet) {
}
