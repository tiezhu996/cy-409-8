package store

import (
	"os"
	"path/filepath"
	"testing"

	"lawsearch/pkg/models"
)

func newTestStore(t *testing.T) (*Store, string) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		s.Close()
		os.Remove(path)
	})
	return s, path
}

func makeBundle(lawID, lawName string, articleCount int) models.ImportBundle {
	bundle := models.ImportBundle{
		Law: models.Law{ID: lawID, Name: lawName},
	}
	for i := 1; i <= articleCount; i++ {
		bundle.Articles = append(bundle.Articles, models.Article{
			ID:      lawID + "-article-" + itoa(i),
			LawID:   lawID,
			Number:  i,
			Title:   "第" + itoa(i) + "条",
			Content: "这是第" + itoa(i) + "条的内容",
		})
	}
	for i := 1; i <= 2; i++ {
		bundle.Sections = append(bundle.Sections, models.Section{
			ID:    lawID + "-section-" + itoa(i),
			LawID: lawID,
			Title: "第" + itoa(i) + "章",
		})
	}
	return bundle
}

func itoa(n int) string {
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if s == "" {
		return "0"
	}
	return s
}

func TestDeleteLaw_ExactMatch(t *testing.T) {
	s, _ := newTestStore(t)

	law1 := makeBundle("民法典", "中华人民共和国民法典", 3)
	law2 := makeBundle("民法典实施条例", "中华人民共和国民法典实施条例", 3)

	if err := s.SaveBundle(law1, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveBundle(law2, nil); err != nil {
		t.Fatal(err)
	}

	articles1, err := s.Articles("民法典")
	if err != nil {
		t.Fatal(err)
	}
	if len(articles1) != 3 {
		t.Fatalf("删除前期望民法典有3条，实际 %d", len(articles1))
	}
	articles2, err := s.Articles("民法典实施条例")
	if err != nil {
		t.Fatal(err)
	}
	if len(articles2) != 3 {
		t.Fatalf("删除前期望实施条例有3条，实际 %d", len(articles2))
	}

	if err := s.DeleteLaw("民法典"); err != nil {
		t.Fatal(err)
	}

	articles1, err = s.Articles("民法典")
	if err != nil {
		t.Fatal(err)
	}
	if len(articles1) != 0 {
		t.Fatalf("删除后期望民法典有0条，实际 %d", len(articles1))
	}
	articles2, err = s.Articles("民法典实施条例")
	if err != nil {
		t.Fatal(err)
	}
	if len(articles2) != 3 {
		t.Fatalf("删除后期望实施条例仍有3条，实际 %d", len(articles2))
	}

	sections1, err := s.Sections("民法典")
	if err != nil {
		t.Fatal(err)
	}
	if len(sections1) != 0 {
		t.Fatalf("删除后期望民法典章节为0，实际 %d", len(sections1))
	}
	sections2, err := s.Sections("民法典实施条例")
	if err != nil {
		t.Fatal(err)
	}
	if len(sections2) != 2 {
		t.Fatalf("删除后期望实施条例章节仍为2，实际 %d", len(sections2))
	}

	laws, err := s.Laws()
	if err != nil {
		t.Fatal(err)
	}
	if len(laws) != 1 {
		t.Fatalf("删除后期望剩余1部法规，实际 %d", len(laws))
	}
	if laws[0].ID != "民法典实施条例" {
		t.Fatalf("剩余法规应为实施条例，实际 %s", laws[0].ID)
	}
}

func TestDeleteLaw_ReimportClearsOldData(t *testing.T) {
	s, _ := newTestStore(t)

	oldBundle := makeBundle("民法典", "中华人民共和国民法典", 5)
	if err := s.SaveBundle(oldBundle, nil); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteLaw("民法典"); err != nil {
		t.Fatal(err)
	}

	newBundle := makeBundle("民法典", "中华人民共和国民法典", 2)
	if err := s.SaveBundle(newBundle, nil); err != nil {
		t.Fatal(err)
	}

	articles, err := s.Articles("民法典")
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 2 {
		t.Fatalf("重导后期望民法典有2条，实际 %d（旧数据未清理干净）", len(articles))
	}

	sections, err := s.Sections("民法典")
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 2 {
		t.Fatalf("重导后期望民法典章节为2，实际 %d", len(sections))
	}
}
