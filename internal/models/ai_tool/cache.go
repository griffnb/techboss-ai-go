package ai_tool

/*
func (this *AiTool) UpdateCache() error {
	err := cache_service.Set(fmt.Sprintf("%s_%s", TABLE, this.ID().String()), this.GetData())
	if err != nil {
		return err
	}
	return nil
}

func GetWithCache(ctx context.Context, id types.UUID) (*AiTool, error) {
	obj, err := New(ctx)
	if err != nil {
		return nil, err
	}

	err = cache_service.Load(fmt.Sprintf("%s_%s", TABLE, id.String()), obj)
	if err != nil {
		return nil, err
	}
	if !tools.Empty(obj) {
		return obj, nil
	}

	obj, err = Get(ctx, id)
	if err != nil {
		return nil, err
	}
	err = obj.UpdateCache()
	if err != nil {
		return nil, err
	}

	return obj, nil
}
*/
