package kepubify

type KepubifyRequestDTO struct {
	CloudToken  string `form:"cloudToken"`
	CloudFolder string `form:"cloudFolder"`
}
