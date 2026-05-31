package format

type KDFInfo struct {
	Name   string `json:"name"`
	N      int    `json:"n"`
	R      int    `json:"r"`
	P      int    `json:"p"`
	KeyLen int    `json:"key_len"`
}

type Metadata struct {
	Version              int     `json:"version"`
	BatchID              string  `json:"batch_id"`
	Salt                 string  `json:"salt"`
	KDF                  KDFInfo `json:"kdf"`
	Shares               int     `json:"shares"`
	Threshold            int     `json:"threshold"`
	OriginalFileSize     int64   `json:"original_file_size"`
	EncryptedPayloadSize int64   `json:"encrypted_payload_size"`
	EncryptedFileName    string  `json:"encrypted_file_name,omitempty"`
	Prefix               string  `json:"prefix"`
	Encrypted            bool    `json:"encrypted"`
	EncryptFileName      bool    `json:"encrypt_file_name"`
}
