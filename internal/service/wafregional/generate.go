//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceARN -ListTagsOutTagsElem=TagInfoForResource.TagList -ServiceTagsSlice -TagInIDElem=ResourceARN -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package wafregional
