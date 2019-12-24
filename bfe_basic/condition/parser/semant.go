package parser

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

var funcProtos = map[string][]Token{
	"default_t":                  nil,
	"req_cip_trusted":            nil,
	"req_vip_in":                 {STRING},
	"req_proto_match":            {STRING},
	"req_proto_secure":           nil,
	"req_host_in":                {STRING},
	"req_host_regmatch":          {STRING},
	"req_path_in":                {STRING, BOOL},
	"req_path_prefix_in":         {STRING, BOOL},
	"req_path_suffix_in":         {STRING, BOOL},
	"req_path_regmatch":          {STRING},
	"req_query_key_prefix_in":    {STRING},
	"req_query_key_in":           {STRING},
	"req_query_exist":            nil,
	"req_query_value_in":         {STRING, STRING, BOOL},
	"req_query_value_prefix_in":  {STRING, STRING, BOOL},
	"req_query_value_suffix_in":  {STRING, STRING, BOOL},
	"req_query_value_regmatch":   {STRING, STRING},
	"req_query_value_contain":    {STRING, STRING, BOOL},
	"req_query_value_hash_in":    {STRING, STRING, BOOL},
	"req_url_regmatch":           {STRING},
	"req_cookie_key_in":          {STRING},
	"req_cookie_value_in":        {STRING, STRING, BOOL},
	"req_cookie_value_prefix_in": {STRING, STRING, BOOL},
	"req_cookie_value_suffix_in": {STRING, STRING, BOOL},
	"req_cookie_value_contain":   {STRING, STRING, BOOL},
	"req_cookie_value_hash_in":   {STRING, STRING, BOOL},
	"req_port_in":                {STRING},
	"req_tag_match":              {STRING, STRING},
	"req_ua_regmatch":            {STRING},
	"req_header_key_in":          {STRING},
	"req_header_value_in":        {STRING, STRING, BOOL},
	"req_header_value_prefix_in": {STRING, STRING, BOOL},
	"req_header_value_suffix_in": {STRING, STRING, BOOL},
	"req_header_value_regmatch":  {STRING, STRING},
	"req_header_value_contain":   {STRING, STRING, BOOL},
	"req_header_value_hash_in":   {STRING, STRING, BOOL},
	"req_method_in":              {STRING},
	"req_cip_range":              {STRING, STRING},
	"req_vip_range":              {STRING, STRING},
	"req_cip_hash_in":            {STRING},
	"res_code_in":                {STRING},
	"res_header_key_in":          {STRING},
	"res_header_value_in":        {STRING, STRING, BOOL},
	"ses_vip_range":              {STRING, STRING},
	"ses_sip_range":              {STRING, STRING},
}

func prototypeCheck(expr *CallExpr) error {
	argsType, ok := funcProtos[expr.Fun.Name]
	if !ok {
		return fmt.Errorf("primitive %s not found", expr.Fun.Name)
	}

	if len(argsType) != len(expr.Args) {
		return fmt.Errorf("primitive args len error, expect %v, got %v", len(argsType), len(expr.Args))
	}

	for i, argType := range argsType {
		if argType != expr.Args[i].Kind {
			return fmt.Errorf("primitive %s arg %d expect %s, got %s", expr.Fun.Name, i, argType, expr.Args[i].Kind)
		}
	}

	return nil
}

func (p *Parser) primitiveCheck(n Node) bool {
	switch x := n.(type) {
	case *BinaryExpr, *UnaryExpr, *ParenExpr:
		return true
	case *Ident:
		return false
	case *CallExpr:
		if err := prototypeCheck(x); err != nil {
			p.addError(x.Pos(), err.Error())
		}
		return false
	default:
		logrus.Printf("get node: %s", n)
	}

	return false
}

func (p *Parser) collectVariable(n Node) bool {
	switch x := n.(type) {
	case *BinaryExpr, *UnaryExpr, *ParenExpr:
		return true
	case *Ident:
		exist := false
		for _, i := range p.identList {
			if i.Name == x.Name {
				exist = true
			}
		}

		if !exist {
			p.identList = append(p.identList, x)
		}
	case *CallExpr:
		return false
	default:
		return false
	}

	return false
}
