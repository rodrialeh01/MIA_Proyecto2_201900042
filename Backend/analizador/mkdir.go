package analizador

type Mkdir struct {
	Path string
	R    bool
}

var consola_mkdir string

func (mkdir *Mkdir) VerificarParams(parametros map[string]string) {
}

func RetornarConsolamkdir() string {
	return consola_mkdir
}
