package util

func AmIMod(myKey string) (bool, error) {
	mods, _ := GetOnlineMods()
	for _, mod := range mods {
		if len(mod.PublicKey) > 0 && mod.PublicKey == myKey {
			return true, nil
		}
	}

	return false, nil
}
