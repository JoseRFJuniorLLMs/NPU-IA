package llm

import (
	"encoding/json"
	"os"
)

// Tokenizer gerencia tokenização de texto
type Tokenizer struct {
	vocab     map[string]int64
	vocabRev  map[int64]string
	eosToken  int64
	padToken  int64
	bosToken  int64
}

// NewTokenizer carrega um tokenizer
func NewTokenizer(path string) (*Tokenizer, error) {
	t := &Tokenizer{
		vocab:    make(map[string]int64),
		vocabRev: make(map[int64]string),
		eosToken: 2,  // Default
		padToken: 0,
		bosToken: 1,
	}

	// Carrega vocabulário se existir
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			var vocab map[string]int64
			if err := json.Unmarshal(data, &vocab); err == nil {
				t.vocab = vocab
				for k, v := range vocab {
					t.vocabRev[v] = k
				}
			}
		}
	}

	return t, nil
}

// Encode converte texto em tokens
func (t *Tokenizer) Encode(text string) ([]int64, []int64) {
	// TODO: Implementar tokenização real (BPE/SentencePiece)
	// Por enquanto, tokenização simples por caractere

	tokens := make([]int64, 0, len(text))
	mask := make([]int64, 0, len(text))

	// Adiciona BOS
	tokens = append(tokens, t.bosToken)
	mask = append(mask, 1)

	for _, char := range text {
		if id, ok := t.vocab[string(char)]; ok {
			tokens = append(tokens, id)
		} else {
			// Token desconhecido
			tokens = append(tokens, 3) // UNK
		}
		mask = append(mask, 1)
	}

	return tokens, mask
}

// Decode converte tokens em texto
func (t *Tokenizer) Decode(tokens []int64) string {
	result := ""
	for _, token := range tokens {
		if token == t.eosToken || token == t.padToken || token == t.bosToken {
			continue
		}
		if str, ok := t.vocabRev[token]; ok {
			result += str
		}
	}
	return result
}

// EOSToken retorna o token de fim de sequência
func (t *Tokenizer) EOSToken() int64 {
	return t.eosToken
}

// PADToken retorna o token de padding
func (t *Tokenizer) PADToken() int64 {
	return t.padToken
}
