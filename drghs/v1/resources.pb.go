// Code generated by protoc-gen-go. DO NOT EDIT.
// source: resources.proto

package drghs_v1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Issue_Priority int32

const (
	Issue_PRIORITY_UNSPECIFIED Issue_Priority = 0
	Issue_P0                   Issue_Priority = 1
	Issue_P1                   Issue_Priority = 2
	Issue_P2                   Issue_Priority = 3
	Issue_P3                   Issue_Priority = 4
	Issue_P4                   Issue_Priority = 5
)

var Issue_Priority_name = map[int32]string{
	0: "PRIORITY_UNSPECIFIED",
	1: "P0",
	2: "P1",
	3: "P2",
	4: "P3",
	5: "P4",
}

var Issue_Priority_value = map[string]int32{
	"PRIORITY_UNSPECIFIED": 0,
	"P0":                   1,
	"P1":                   2,
	"P2":                   3,
	"P3":                   4,
	"P4":                   5,
}

func (x Issue_Priority) String() string {
	return proto.EnumName(Issue_Priority_name, int32(x))
}

func (Issue_Priority) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{4, 0}
}

type Issue_IssueType int32

const (
	Issue_GITHUB_ISSUE_TYPE_UNSPECIFIED Issue_IssueType = 0
	Issue_BUG                           Issue_IssueType = 1
	Issue_FEATURE                       Issue_IssueType = 2
	Issue_QUESTION                      Issue_IssueType = 3
	Issue_CLEANUP                       Issue_IssueType = 4
	Issue_PROCESS                       Issue_IssueType = 5
)

var Issue_IssueType_name = map[int32]string{
	0: "GITHUB_ISSUE_TYPE_UNSPECIFIED",
	1: "BUG",
	2: "FEATURE",
	3: "QUESTION",
	4: "CLEANUP",
	5: "PROCESS",
}

var Issue_IssueType_value = map[string]int32{
	"GITHUB_ISSUE_TYPE_UNSPECIFIED": 0,
	"BUG":                           1,
	"FEATURE":                       2,
	"QUESTION":                      3,
	"CLEANUP":                       4,
	"PROCESS":                       5,
}

func (x Issue_IssueType) String() string {
	return proto.EnumName(Issue_IssueType_name, int32(x))
}

func (Issue_IssueType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{4, 1}
}

type Repository struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Repository) Reset()         { *m = Repository{} }
func (m *Repository) String() string { return proto.CompactTextString(m) }
func (*Repository) ProtoMessage()    {}
func (*Repository) Descriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{0}
}

func (m *Repository) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Repository.Unmarshal(m, b)
}
func (m *Repository) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Repository.Marshal(b, m, deterministic)
}
func (m *Repository) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Repository.Merge(m, src)
}
func (m *Repository) XXX_Size() int {
	return xxx_messageInfo_Repository.Size(m)
}
func (m *Repository) XXX_DiscardUnknown() {
	xxx_messageInfo_Repository.DiscardUnknown(m)
}

var xxx_messageInfo_Repository proto.InternalMessageInfo

func (m *Repository) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type GitCommit struct {
	Name                 string               `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Subject              string               `protobuf:"bytes,2,opt,name=subject,proto3" json:"subject,omitempty"`
	AuthorEmail          string               `protobuf:"bytes,3,opt,name=author_email,json=authorEmail,proto3" json:"author_email,omitempty"`
	AuthoredTime         *timestamp.Timestamp `protobuf:"bytes,4,opt,name=authored_time,json=authoredTime,proto3" json:"authored_time,omitempty"`
	CommitterEmail       string               `protobuf:"bytes,5,opt,name=committer_email,json=committerEmail,proto3" json:"committer_email,omitempty"`
	CommittedTime        *timestamp.Timestamp `protobuf:"bytes,6,opt,name=committed_time,json=committedTime,proto3" json:"committed_time,omitempty"`
	Sha                  string               `protobuf:"bytes,7,opt,name=sha,proto3" json:"sha,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *GitCommit) Reset()         { *m = GitCommit{} }
func (m *GitCommit) String() string { return proto.CompactTextString(m) }
func (*GitCommit) ProtoMessage()    {}
func (*GitCommit) Descriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{1}
}

func (m *GitCommit) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GitCommit.Unmarshal(m, b)
}
func (m *GitCommit) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GitCommit.Marshal(b, m, deterministic)
}
func (m *GitCommit) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GitCommit.Merge(m, src)
}
func (m *GitCommit) XXX_Size() int {
	return xxx_messageInfo_GitCommit.Size(m)
}
func (m *GitCommit) XXX_DiscardUnknown() {
	xxx_messageInfo_GitCommit.DiscardUnknown(m)
}

var xxx_messageInfo_GitCommit proto.InternalMessageInfo

func (m *GitCommit) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *GitCommit) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *GitCommit) GetAuthorEmail() string {
	if m != nil {
		return m.AuthorEmail
	}
	return ""
}

func (m *GitCommit) GetAuthoredTime() *timestamp.Timestamp {
	if m != nil {
		return m.AuthoredTime
	}
	return nil
}

func (m *GitCommit) GetCommitterEmail() string {
	if m != nil {
		return m.CommitterEmail
	}
	return ""
}

func (m *GitCommit) GetCommittedTime() *timestamp.Timestamp {
	if m != nil {
		return m.CommittedTime
	}
	return nil
}

func (m *GitCommit) GetSha() string {
	if m != nil {
		return m.Sha
	}
	return ""
}

type GitHubUser struct {
	Id                   int32    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Login                string   `protobuf:"bytes,2,opt,name=login,proto3" json:"login,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GitHubUser) Reset()         { *m = GitHubUser{} }
func (m *GitHubUser) String() string { return proto.CompactTextString(m) }
func (*GitHubUser) ProtoMessage()    {}
func (*GitHubUser) Descriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{2}
}

func (m *GitHubUser) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GitHubUser.Unmarshal(m, b)
}
func (m *GitHubUser) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GitHubUser.Marshal(b, m, deterministic)
}
func (m *GitHubUser) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GitHubUser.Merge(m, src)
}
func (m *GitHubUser) XXX_Size() int {
	return xxx_messageInfo_GitHubUser.Size(m)
}
func (m *GitHubUser) XXX_DiscardUnknown() {
	xxx_messageInfo_GitHubUser.DiscardUnknown(m)
}

var xxx_messageInfo_GitHubUser proto.InternalMessageInfo

func (m *GitHubUser) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *GitHubUser) GetLogin() string {
	if m != nil {
		return m.Login
	}
	return ""
}

type GitHubComment struct {
	Id                   int32                `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	User                 *GitHubUser          `protobuf:"bytes,2,opt,name=user,proto3" json:"user,omitempty"`
	CreatedAt            *timestamp.Timestamp `protobuf:"bytes,3,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt            *timestamp.Timestamp `protobuf:"bytes,4,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Body                 string               `protobuf:"bytes,5,opt,name=body,proto3" json:"body,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *GitHubComment) Reset()         { *m = GitHubComment{} }
func (m *GitHubComment) String() string { return proto.CompactTextString(m) }
func (*GitHubComment) ProtoMessage()    {}
func (*GitHubComment) Descriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{3}
}

func (m *GitHubComment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GitHubComment.Unmarshal(m, b)
}
func (m *GitHubComment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GitHubComment.Marshal(b, m, deterministic)
}
func (m *GitHubComment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GitHubComment.Merge(m, src)
}
func (m *GitHubComment) XXX_Size() int {
	return xxx_messageInfo_GitHubComment.Size(m)
}
func (m *GitHubComment) XXX_DiscardUnknown() {
	xxx_messageInfo_GitHubComment.DiscardUnknown(m)
}

var xxx_messageInfo_GitHubComment proto.InternalMessageInfo

func (m *GitHubComment) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *GitHubComment) GetUser() *GitHubUser {
	if m != nil {
		return m.User
	}
	return nil
}

func (m *GitHubComment) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *GitHubComment) GetUpdatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.UpdatedAt
	}
	return nil
}

func (m *GitHubComment) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

type Issue struct {
	Name                 string               `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Title                string               `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Body                 string               `protobuf:"bytes,3,opt,name=body,proto3" json:"body,omitempty"`
	Priority             Issue_Priority       `protobuf:"varint,4,opt,name=priority,proto3,enum=drghs.v1.Issue_Priority" json:"priority,omitempty"`
	IssueType            Issue_IssueType      `protobuf:"varint,5,opt,name=issue_type,json=issueType,proto3,enum=drghs.v1.Issue_IssueType" json:"issue_type,omitempty"`
	Labels               []string             `protobuf:"bytes,6,rep,name=labels,proto3" json:"labels,omitempty"`
	CreatedAt            *timestamp.Timestamp `protobuf:"bytes,7,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt            *timestamp.Timestamp `protobuf:"bytes,8,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	ClosedAt             *timestamp.Timestamp `protobuf:"bytes,9,opt,name=closed_at,json=closedAt,proto3" json:"closed_at,omitempty"`
	Closed               bool                 `protobuf:"varint,10,opt,name=closed,proto3" json:"closed,omitempty"`
	ClosedBy             *GitHubUser          `protobuf:"bytes,11,opt,name=closed_by,json=closedBy,proto3" json:"closed_by,omitempty"`
	IsPr                 bool                 `protobuf:"varint,12,opt,name=is_pr,json=isPr,proto3" json:"is_pr,omitempty"`
	Approved             bool                 `protobuf:"varint,13,opt,name=approved,proto3" json:"approved,omitempty"`
	GitCommit            *GitCommit           `protobuf:"bytes,14,opt,name=git_commit,json=gitCommit,proto3" json:"git_commit,omitempty"`
	IssueId              int32                `protobuf:"varint,15,opt,name=issue_id,json=issueId,proto3" json:"issue_id,omitempty"`
	Url                  string               `protobuf:"bytes,16,opt,name=url,proto3" json:"url,omitempty"`
	Assignees            []*GitHubUser        `protobuf:"bytes,17,rep,name=assignees,proto3" json:"assignees,omitempty"`
	Reporter             *GitHubUser          `protobuf:"bytes,18,opt,name=reporter,proto3" json:"reporter,omitempty"`
	Comments             []*GitHubComment     `protobuf:"bytes,19,rep,name=comments,proto3" json:"comments,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Issue) Reset()         { *m = Issue{} }
func (m *Issue) String() string { return proto.CompactTextString(m) }
func (*Issue) ProtoMessage()    {}
func (*Issue) Descriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{4}
}

func (m *Issue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Issue.Unmarshal(m, b)
}
func (m *Issue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Issue.Marshal(b, m, deterministic)
}
func (m *Issue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Issue.Merge(m, src)
}
func (m *Issue) XXX_Size() int {
	return xxx_messageInfo_Issue.Size(m)
}
func (m *Issue) XXX_DiscardUnknown() {
	xxx_messageInfo_Issue.DiscardUnknown(m)
}

var xxx_messageInfo_Issue proto.InternalMessageInfo

func (m *Issue) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Issue) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *Issue) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

func (m *Issue) GetPriority() Issue_Priority {
	if m != nil {
		return m.Priority
	}
	return Issue_PRIORITY_UNSPECIFIED
}

func (m *Issue) GetIssueType() Issue_IssueType {
	if m != nil {
		return m.IssueType
	}
	return Issue_GITHUB_ISSUE_TYPE_UNSPECIFIED
}

func (m *Issue) GetLabels() []string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *Issue) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *Issue) GetUpdatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.UpdatedAt
	}
	return nil
}

func (m *Issue) GetClosedAt() *timestamp.Timestamp {
	if m != nil {
		return m.ClosedAt
	}
	return nil
}

func (m *Issue) GetClosed() bool {
	if m != nil {
		return m.Closed
	}
	return false
}

func (m *Issue) GetClosedBy() *GitHubUser {
	if m != nil {
		return m.ClosedBy
	}
	return nil
}

func (m *Issue) GetIsPr() bool {
	if m != nil {
		return m.IsPr
	}
	return false
}

func (m *Issue) GetApproved() bool {
	if m != nil {
		return m.Approved
	}
	return false
}

func (m *Issue) GetGitCommit() *GitCommit {
	if m != nil {
		return m.GitCommit
	}
	return nil
}

func (m *Issue) GetIssueId() int32 {
	if m != nil {
		return m.IssueId
	}
	return 0
}

func (m *Issue) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *Issue) GetAssignees() []*GitHubUser {
	if m != nil {
		return m.Assignees
	}
	return nil
}

func (m *Issue) GetReporter() *GitHubUser {
	if m != nil {
		return m.Reporter
	}
	return nil
}

func (m *Issue) GetComments() []*GitHubComment {
	if m != nil {
		return m.Comments
	}
	return nil
}

type File struct {
	// Output only. The full path of the  [File][] within its [Repository][].
	Filepath string `protobuf:"bytes,1,opt,name=filepath,proto3" json:"filepath,omitempty"`
	// Output only. The [GitCommit][] of the file.
	GitCommit            *GitCommit `protobuf:"bytes,2,opt,name=git_commit,json=gitCommit,proto3" json:"git_commit,omitempty"`
	Size                 int32      `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *File) Reset()         { *m = File{} }
func (m *File) String() string { return proto.CompactTextString(m) }
func (*File) ProtoMessage()    {}
func (*File) Descriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{5}
}

func (m *File) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_File.Unmarshal(m, b)
}
func (m *File) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_File.Marshal(b, m, deterministic)
}
func (m *File) XXX_Merge(src proto.Message) {
	xxx_messageInfo_File.Merge(m, src)
}
func (m *File) XXX_Size() int {
	return xxx_messageInfo_File.Size(m)
}
func (m *File) XXX_DiscardUnknown() {
	xxx_messageInfo_File.DiscardUnknown(m)
}

var xxx_messageInfo_File proto.InternalMessageInfo

func (m *File) GetFilepath() string {
	if m != nil {
		return m.Filepath
	}
	return ""
}

func (m *File) GetGitCommit() *GitCommit {
	if m != nil {
		return m.GitCommit
	}
	return nil
}

func (m *File) GetSize() int32 {
	if m != nil {
		return m.Size
	}
	return 0
}

type SnippetVersionMeta struct {
	// Output only. Used as metadata information on the [SnipeptVersion[
	Title                string   `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description          string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Usage                string   `protobuf:"bytes,3,opt,name=usage,proto3" json:"usage,omitempty"`
	ApiVersion           string   `protobuf:"bytes,4,opt,name=api_version,json=apiVersion,proto3" json:"api_version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SnippetVersionMeta) Reset()         { *m = SnippetVersionMeta{} }
func (m *SnippetVersionMeta) String() string { return proto.CompactTextString(m) }
func (*SnippetVersionMeta) ProtoMessage()    {}
func (*SnippetVersionMeta) Descriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{6}
}

func (m *SnippetVersionMeta) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SnippetVersionMeta.Unmarshal(m, b)
}
func (m *SnippetVersionMeta) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SnippetVersionMeta.Marshal(b, m, deterministic)
}
func (m *SnippetVersionMeta) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnippetVersionMeta.Merge(m, src)
}
func (m *SnippetVersionMeta) XXX_Size() int {
	return xxx_messageInfo_SnippetVersionMeta.Size(m)
}
func (m *SnippetVersionMeta) XXX_DiscardUnknown() {
	xxx_messageInfo_SnippetVersionMeta.DiscardUnknown(m)
}

var xxx_messageInfo_SnippetVersionMeta proto.InternalMessageInfo

func (m *SnippetVersionMeta) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *SnippetVersionMeta) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *SnippetVersionMeta) GetUsage() string {
	if m != nil {
		return m.Usage
	}
	return ""
}

func (m *SnippetVersionMeta) GetApiVersion() string {
	if m != nil {
		return m.ApiVersion
	}
	return ""
}

type SnippetVersion struct {
	// Output only. The resource name for the [SnippetVersion][] in the format
	// `owners/*/repositories/*/snippets/*/snippetVersions/*`.
	Name                 string              `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	File                 *File               `protobuf:"bytes,2,opt,name=file,proto3" json:"file,omitempty"`
	Lines                []string            `protobuf:"bytes,3,rep,name=lines,proto3" json:"lines,omitempty"`
	Content              string              `protobuf:"bytes,4,opt,name=content,proto3" json:"content,omitempty"`
	Meta                 *SnippetVersionMeta `protobuf:"bytes,5,opt,name=meta,proto3" json:"meta,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *SnippetVersion) Reset()         { *m = SnippetVersion{} }
func (m *SnippetVersion) String() string { return proto.CompactTextString(m) }
func (*SnippetVersion) ProtoMessage()    {}
func (*SnippetVersion) Descriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{7}
}

func (m *SnippetVersion) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SnippetVersion.Unmarshal(m, b)
}
func (m *SnippetVersion) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SnippetVersion.Marshal(b, m, deterministic)
}
func (m *SnippetVersion) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnippetVersion.Merge(m, src)
}
func (m *SnippetVersion) XXX_Size() int {
	return xxx_messageInfo_SnippetVersion.Size(m)
}
func (m *SnippetVersion) XXX_DiscardUnknown() {
	xxx_messageInfo_SnippetVersion.DiscardUnknown(m)
}

var xxx_messageInfo_SnippetVersion proto.InternalMessageInfo

func (m *SnippetVersion) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *SnippetVersion) GetFile() *File {
	if m != nil {
		return m.File
	}
	return nil
}

func (m *SnippetVersion) GetLines() []string {
	if m != nil {
		return m.Lines
	}
	return nil
}

func (m *SnippetVersion) GetContent() string {
	if m != nil {
		return m.Content
	}
	return ""
}

func (m *SnippetVersion) GetMeta() *SnippetVersionMeta {
	if m != nil {
		return m.Meta
	}
	return nil
}

type Snippet struct {
	// Output only. The resource name for the [Snippet][] in the format
	// `owners/*/repositories/*/snippets/*/languages/*`.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Output only. The programming language of the snippet.
	// TODO(jdobry): Switch this from a string to an enum of the languages from
	// https://github.com/src-d/enry.
	Language string `protobuf:"bytes,2,opt,name=language,proto3" json:"language,omitempty"`
	// Output only. A copy of the most recent [SnippetVersion][] of the
	// [Snippet][].
	Primary              *SnippetVersion `protobuf:"bytes,3,opt,name=primary,proto3" json:"primary,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Snippet) Reset()         { *m = Snippet{} }
func (m *Snippet) String() string { return proto.CompactTextString(m) }
func (*Snippet) ProtoMessage()    {}
func (*Snippet) Descriptor() ([]byte, []int) {
	return fileDescriptor_cf1b13971fe4c19d, []int{8}
}

func (m *Snippet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Snippet.Unmarshal(m, b)
}
func (m *Snippet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Snippet.Marshal(b, m, deterministic)
}
func (m *Snippet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Snippet.Merge(m, src)
}
func (m *Snippet) XXX_Size() int {
	return xxx_messageInfo_Snippet.Size(m)
}
func (m *Snippet) XXX_DiscardUnknown() {
	xxx_messageInfo_Snippet.DiscardUnknown(m)
}

var xxx_messageInfo_Snippet proto.InternalMessageInfo

func (m *Snippet) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Snippet) GetLanguage() string {
	if m != nil {
		return m.Language
	}
	return ""
}

func (m *Snippet) GetPrimary() *SnippetVersion {
	if m != nil {
		return m.Primary
	}
	return nil
}

func init() {
	proto.RegisterEnum("drghs.v1.Issue_Priority", Issue_Priority_name, Issue_Priority_value)
	proto.RegisterEnum("drghs.v1.Issue_IssueType", Issue_IssueType_name, Issue_IssueType_value)
	proto.RegisterType((*Repository)(nil), "drghs.v1.Repository")
	proto.RegisterType((*GitCommit)(nil), "drghs.v1.GitCommit")
	proto.RegisterType((*GitHubUser)(nil), "drghs.v1.GitHubUser")
	proto.RegisterType((*GitHubComment)(nil), "drghs.v1.GitHubComment")
	proto.RegisterType((*Issue)(nil), "drghs.v1.Issue")
	proto.RegisterType((*File)(nil), "drghs.v1.File")
	proto.RegisterType((*SnippetVersionMeta)(nil), "drghs.v1.SnippetVersionMeta")
	proto.RegisterType((*SnippetVersion)(nil), "drghs.v1.SnippetVersion")
	proto.RegisterType((*Snippet)(nil), "drghs.v1.Snippet")
}

func init() { proto.RegisterFile("resources.proto", fileDescriptor_cf1b13971fe4c19d) }

var fileDescriptor_cf1b13971fe4c19d = []byte{
	// 951 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x55, 0xcf, 0x6e, 0xdb, 0xc6,
	0x13, 0xfe, 0x51, 0xa2, 0x2c, 0x72, 0x64, 0xcb, 0xfc, 0xad, 0x8d, 0x76, 0x63, 0xb4, 0x88, 0xc2,
	0x4b, 0x75, 0x52, 0x6c, 0x3a, 0x40, 0xdb, 0x53, 0x21, 0xbb, 0xb4, 0x23, 0x20, 0xb5, 0xd5, 0x95,
	0x54, 0x20, 0x27, 0x81, 0x12, 0x37, 0xf2, 0x06, 0x14, 0x49, 0xec, 0x2e, 0x0d, 0xa8, 0xd7, 0x3e,
	0x46, 0xdf, 0xa0, 0xa7, 0xbe, 0x4b, 0x5f, 0xa8, 0xe0, 0xf0, 0x8f, 0x12, 0xc5, 0xb1, 0xdb, 0x5e,
	0xc4, 0xfd, 0x66, 0x67, 0xe6, 0x9b, 0xf9, 0x76, 0x77, 0x04, 0x87, 0x92, 0xab, 0x24, 0x93, 0x4b,
	0xae, 0x06, 0xa9, 0x4c, 0x74, 0x42, 0xac, 0x50, 0xae, 0xee, 0xd4, 0xe0, 0xfe, 0xec, 0xe4, 0xf9,
	0x2a, 0x49, 0x56, 0x11, 0x7f, 0x89, 0xf6, 0x45, 0xf6, 0xee, 0xa5, 0x16, 0x6b, 0xae, 0x74, 0xb0,
	0x4e, 0x0b, 0x57, 0xb7, 0x07, 0xc0, 0x78, 0x9a, 0x28, 0xa1, 0x13, 0xb9, 0x21, 0x04, 0xcc, 0x38,
	0x58, 0x73, 0x6a, 0xf4, 0x8c, 0xbe, 0xcd, 0x70, 0xed, 0xfe, 0xde, 0x00, 0xfb, 0x5a, 0xe8, 0xcb,
	0x64, 0xbd, 0x16, 0xfa, 0x21, 0x0f, 0x42, 0xa1, 0xad, 0xb2, 0xc5, 0x7b, 0xbe, 0xd4, 0xb4, 0x81,
	0xe6, 0x0a, 0x92, 0x17, 0xb0, 0x1f, 0x64, 0xfa, 0x2e, 0x91, 0x73, 0xbe, 0x0e, 0x44, 0x44, 0x9b,
	0xb8, 0xdd, 0x29, 0x6c, 0x7e, 0x6e, 0x22, 0x3f, 0xc0, 0x41, 0x01, 0x79, 0x38, 0xcf, 0x8b, 0xa3,
	0x66, 0xcf, 0xe8, 0x77, 0xbc, 0x93, 0x41, 0x51, 0xf9, 0xa0, 0xaa, 0x7c, 0x30, 0xad, 0x2a, 0x67,
	0xfb, 0x55, 0x40, 0x6e, 0x22, 0xdf, 0xc0, 0xe1, 0x12, 0x6b, 0xd3, 0xbc, 0xa2, 0x69, 0x21, 0x4d,
	0xb7, 0x36, 0x17, 0x4c, 0x43, 0xa8, 0x2d, 0x25, 0xd5, 0xde, 0x93, 0x54, 0x07, 0x75, 0x04, 0x72,
	0x39, 0xd0, 0x54, 0x77, 0x01, 0x6d, 0x63, 0xfe, 0x7c, 0xe9, 0x7a, 0x00, 0xd7, 0x42, 0xbf, 0xce,
	0x16, 0x33, 0xc5, 0x25, 0xe9, 0x42, 0x43, 0x84, 0xa8, 0x4d, 0x8b, 0x35, 0x44, 0x48, 0x8e, 0xa1,
	0x15, 0x25, 0x2b, 0x11, 0x97, 0xba, 0x14, 0xc0, 0xfd, 0xcb, 0x80, 0x83, 0x22, 0x28, 0x17, 0x95,
	0xc7, 0xfa, 0x93, 0xb8, 0x3e, 0x98, 0x99, 0xe2, 0x12, 0xc3, 0x3a, 0xde, 0xf1, 0xa0, 0x3a, 0xcf,
	0xc1, 0x96, 0x8b, 0xa1, 0x07, 0xf9, 0x1e, 0x60, 0x29, 0x79, 0x90, 0xb7, 0x14, 0x68, 0xd4, 0xf7,
	0xf1, 0x86, 0xec, 0xd2, 0x7b, 0xa8, 0xf3, 0xd0, 0x2c, 0x0d, 0xab, 0xd0, 0xa7, 0x65, 0xb7, 0x4b,
	0xef, 0x21, 0xde, 0x82, 0x45, 0x12, 0x6e, 0x4a, 0xa1, 0x71, 0xed, 0xfe, 0xd9, 0x86, 0xd6, 0x48,
	0xa9, 0x8c, 0x3f, 0x78, 0x47, 0x8e, 0xa1, 0xa5, 0x85, 0x8e, 0x78, 0xa5, 0x04, 0x82, 0x3a, 0x4f,
	0x73, 0x9b, 0x87, 0xbc, 0x02, 0x2b, 0x95, 0x22, 0x91, 0x42, 0x6f, 0xb0, 0xa8, 0xae, 0x47, 0xb7,
	0xfd, 0x23, 0xc1, 0x60, 0x5c, 0xee, 0xb3, 0xda, 0x93, 0x7c, 0x07, 0x20, 0xf2, 0xbd, 0xb9, 0xde,
	0xa4, 0x1c, 0xeb, 0xea, 0x7a, 0xcf, 0x76, 0xe3, 0xf0, 0x77, 0xba, 0x49, 0x39, 0xb3, 0x45, 0xb5,
	0x24, 0x5f, 0xc0, 0x5e, 0x14, 0x2c, 0x78, 0xa4, 0xe8, 0x5e, 0xaf, 0xd9, 0xb7, 0x59, 0x89, 0x76,
	0x94, 0x6d, 0xff, 0x77, 0x65, 0xad, 0x7f, 0xa3, 0xec, 0xb7, 0x60, 0x2f, 0xa3, 0x44, 0x15, 0x91,
	0xf6, 0x93, 0x91, 0x56, 0xe1, 0x3c, 0xd4, 0x79, 0x1b, 0xc5, 0x9a, 0x42, 0xcf, 0xe8, 0x5b, 0xac,
	0x44, 0xe4, 0xac, 0x4e, 0xb8, 0xd8, 0xd0, 0xce, 0x23, 0xf7, 0xa9, 0x4c, 0x75, 0xb1, 0x21, 0x47,
	0xd0, 0x12, 0x6a, 0x9e, 0x4a, 0xba, 0x8f, 0x99, 0x4c, 0xa1, 0xc6, 0x92, 0x9c, 0x80, 0x15, 0xa4,
	0xa9, 0x4c, 0xee, 0x79, 0x48, 0x0f, 0xd0, 0x5e, 0x63, 0xe2, 0x01, 0xac, 0x84, 0x9e, 0x17, 0x6f,
	0x85, 0x76, 0x91, 0xe4, 0xe8, 0x23, 0x92, 0x62, 0x7a, 0x30, 0x7b, 0x55, 0x0f, 0x92, 0x67, 0x60,
	0x15, 0x07, 0x26, 0x42, 0x7a, 0x88, 0x17, 0xbf, 0x8d, 0x78, 0x14, 0xe6, 0xaf, 0x2c, 0x93, 0x11,
	0x75, 0x8a, 0x57, 0x96, 0xc9, 0x88, 0x78, 0x60, 0x07, 0x4a, 0x89, 0x55, 0xcc, 0xb9, 0xa2, 0xff,
	0xef, 0x35, 0x3f, 0xdb, 0xc4, 0xd6, 0x8d, 0x9c, 0x82, 0x25, 0x79, 0x9a, 0x48, 0xcd, 0x25, 0x25,
	0x8f, 0xf5, 0x5d, 0x79, 0x91, 0x73, 0xb0, 0x96, 0xc5, 0x83, 0x54, 0xf4, 0x08, 0x49, 0xbe, 0xdc,
	0x8d, 0x28, 0x1f, 0x2c, 0xab, 0x1d, 0xdd, 0x37, 0x60, 0x55, 0xd7, 0x91, 0x50, 0x38, 0x1e, 0xb3,
	0xd1, 0x2d, 0x1b, 0x4d, 0xdf, 0xce, 0x67, 0x37, 0x93, 0xb1, 0x7f, 0x39, 0xba, 0x1a, 0xf9, 0x3f,
	0x3a, 0xff, 0x23, 0x7b, 0xd0, 0x18, 0x9f, 0x3a, 0x06, 0x7e, 0xcf, 0x9c, 0x06, 0x7e, 0x3d, 0xa7,
	0x89, 0xdf, 0x73, 0xc7, 0xc4, 0xef, 0x2b, 0xa7, 0xe5, 0x46, 0x60, 0xd7, 0x97, 0x94, 0xbc, 0x80,
	0xaf, 0xaf, 0x47, 0xd3, 0xd7, 0xb3, 0x8b, 0xf9, 0x68, 0x32, 0x99, 0xf9, 0xf3, 0xe9, 0xdb, 0xb1,
	0xbf, 0x93, 0xb7, 0x0d, 0xcd, 0x8b, 0xd9, 0xb5, 0x63, 0x90, 0x0e, 0xb4, 0xaf, 0xfc, 0xe1, 0x74,
	0xc6, 0x7c, 0xa7, 0x41, 0xf6, 0xc1, 0xfa, 0x79, 0xe6, 0x4f, 0xa6, 0xa3, 0xdb, 0x1b, 0xa7, 0x99,
	0x6f, 0x5d, 0xbe, 0xf1, 0x87, 0x37, 0xb3, 0xb1, 0x63, 0xe6, 0x60, 0xcc, 0x6e, 0x2f, 0xfd, 0xc9,
	0xc4, 0x69, 0xb9, 0xef, 0xc1, 0xbc, 0x12, 0x11, 0xcf, 0xcf, 0xf6, 0x9d, 0x88, 0x78, 0x1a, 0xe8,
	0xbb, 0xf2, 0xd1, 0xd6, 0x78, 0xe7, 0x6c, 0x1b, 0xff, 0xe8, 0x6c, 0x09, 0x98, 0x4a, 0xfc, 0xca,
	0xf1, 0x59, 0xb7, 0x18, 0xae, 0xdd, 0xdf, 0x0c, 0x20, 0x93, 0x58, 0xa4, 0x29, 0xd7, 0xbf, 0x70,
	0xa9, 0x44, 0x12, 0xff, 0xc4, 0x75, 0xb0, 0x9d, 0x0b, 0xc6, 0x87, 0x73, 0xa1, 0x07, 0x9d, 0x90,
	0xab, 0xa5, 0x14, 0xa9, 0x16, 0x49, 0x35, 0x3d, 0x3f, 0x34, 0xe5, 0x71, 0x99, 0x0a, 0x56, 0xbc,
	0x1c, 0x1d, 0x05, 0x20, 0xcf, 0xa1, 0x13, 0xa4, 0x62, 0x7e, 0x5f, 0x10, 0xe0, 0xf8, 0xb0, 0x19,
	0x04, 0xa9, 0x28, 0x29, 0xdd, 0x3f, 0x0c, 0xe8, 0x7e, 0x5c, 0xc5, 0x83, 0xd3, 0xca, 0x05, 0x33,
	0x17, 0xa0, 0x6c, 0xb7, 0xbb, 0x6d, 0x37, 0x97, 0x8b, 0xe1, 0x1e, 0xce, 0x76, 0x11, 0x73, 0x45,
	0x9b, 0x38, 0x36, 0x0a, 0x90, 0xff, 0x17, 0x2e, 0x93, 0x58, 0xf3, 0x58, 0x97, 0xec, 0x15, 0x24,
	0xa7, 0x60, 0xae, 0xb9, 0x0e, 0x70, 0x36, 0x75, 0xbc, 0xaf, 0xb6, 0x39, 0x3f, 0x55, 0x85, 0xa1,
	0xa7, 0xbb, 0x86, 0x76, 0xb9, 0xf7, 0x60, 0x91, 0x27, 0x60, 0x45, 0x41, 0xbc, 0xca, 0x72, 0x15,
	0x0a, 0x85, 0x6a, 0x4c, 0x3c, 0x68, 0xa7, 0x52, 0xac, 0x03, 0xb9, 0x29, 0xff, 0x13, 0xe8, 0xe7,
	0xf8, 0x58, 0xe5, 0xb8, 0xd8, 0xc3, 0xf9, 0x72, 0xfe, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x92,
	0x03, 0x27, 0x18, 0x4f, 0x08, 0x00, 0x00,
}
