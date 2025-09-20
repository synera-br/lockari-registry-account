package utils

func Contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// comparable é uma interface pré-declarada em Go que define um conjunto de tipos que podem ser comparados
// usando os operadores == e !=.
//
// Na declaração `func ContainsGeneric[T comparable]`, estamos dizendo que a função `ContainsGeneric`
// é genérica e aceita qualquer tipo `T` que implemente a interface `comparable`. Isso significa que o tipo `T`
// deve suportar comparação usando == e !=.
func ContainsGeneric[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
