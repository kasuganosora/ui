package widget

// Emoji is one handwritten catalog entry. Labels are original short text for
// this repository, not an emojibase keyword dump.
type Emoji struct {
	Char     string
	Category string
	Label    string
}

// EmojiCatalog is the built-in search set. Skin tones and ZWJ sequences are
// limited to the rows written here.
func EmojiCatalog() []Emoji {
	return []Emoji{
		{Char: "😀", Category: "表情", Label: "笑脸 grin"},
		{Char: "😁", Category: "表情", Label: "露齿 beam"},
		{Char: "😂", Category: "表情", Label: "笑哭 joy"},
		{Char: "🤣", Category: "表情", Label: "打滚 rofl"},
		{Char: "😊", Category: "表情", Label: "微笑 blush"},
		{Char: "😍", Category: "表情", Label: "心眼 heart"},
		{Char: "😘", Category: "表情", Label: "飞吻 kiss"},
		{Char: "😎", Category: "表情", Label: "墨镜 cool"},
		{Char: "🤔", Category: "表情", Label: "思考 think"},
		{Char: "😴", Category: "表情", Label: "睡觉 sleep"},
		{Char: "😢", Category: "表情", Label: "哭泣 cry"},
		{Char: "😡", Category: "表情", Label: "生气 angry"},
		{Char: "👍", Category: "手势", Label: "点赞 thumb"},
		{Char: "👎", Category: "手势", Label: "点踩 down"},
		{Char: "👏", Category: "手势", Label: "鼓掌 clap"},
		{Char: "🙏", Category: "手势", Label: "合十 pray"},
		{Char: "👋", Category: "手势", Label: "挥手 wave"},
		{Char: "✌️", Category: "手势", Label: "胜利 peace"},
		{Char: "👌", Category: "手势", Label: "好的 ok"},
		{Char: "✊", Category: "手势", Label: "拳头 fist"},
		{Char: "🐶", Category: "物品", Label: "小狗 dog"},
		{Char: "🐱", Category: "物品", Label: "小猫 cat"},
		{Char: "🍎", Category: "物品", Label: "苹果 apple"},
		{Char: "☕", Category: "物品", Label: "咖啡 coffee"},
		{Char: "🌙", Category: "物品", Label: "月亮 moon"},
		{Char: "⭐", Category: "物品", Label: "星星 star"},
		{Char: "🔥", Category: "物品", Label: "火焰 fire"},
		{Char: "🌈", Category: "物品", Label: "彩虹 rainbow"},
		{Char: "🎉", Category: "物品", Label: "庆祝 party"},
		{Char: "🎁", Category: "物品", Label: "礼物 gift"},
		{Char: "📚", Category: "物品", Label: "书籍 book"},
		{Char: "✏️", Category: "物品", Label: "铅笔 pencil"},
		{Char: "⚽", Category: "物品", Label: "足球 ball"},
		{Char: "🚗", Category: "物品", Label: "汽车 car"},
		{Char: "✈️", Category: "物品", Label: "飞机 plane"},
		{Char: "🏠", Category: "物品", Label: "房子 house"},
	}
}
