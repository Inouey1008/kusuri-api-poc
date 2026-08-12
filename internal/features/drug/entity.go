// Package drug は医薬品モジュールのドメインモデル・永続化・ユースケース・HTTP ハンドラをまとめる。
// モジュールを追加するときは、このパッケージと同じ構成のディレクトリを新設する。
package drug

// 医薬品のドメインエンティティ (JSON タグは dto.go 側で付与する)
type Drug struct {
	YJCode string
	Name   string
}
