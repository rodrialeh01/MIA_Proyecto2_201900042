package analizador

type Mkuser struct {
	User string
	Pwd  string
	Grp  string
}

func (mkuser *Mkuser) VerificarParams(parametros map[string]string) {

}
