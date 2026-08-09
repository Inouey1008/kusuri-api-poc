// Package drug は医薬品モジュールのドメインモデル・永続化・ユースケース・HTTP ハンドラをまとめる。
// モジュールを追加するときは、このパッケージと同じ構成のディレクトリを新設する。
package drug

// Drug は医薬品を表すドメインエンティティ。
// JSON タグは handler.go の DTO 側で付与し、ドメインと転送表現を分離する。
type Drug struct {
	YJCode string
	Name   string
}
