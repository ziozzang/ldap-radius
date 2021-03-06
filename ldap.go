package main

import (
	"crypto/tls"
	"log"
	"strings"

	"gopkg.in/ldap.v2"
)

var ldapConn *ldap.Conn

func ldapLogin(username, password string) bool {
	ldapConn, err := ldap.Dial("tcp", config.Ldap.Host)
	check(err, "could not connect to ldap server")
	defer ldapConn.Close()

	if config.Ldap.Secure {
		err = ldapConn.StartTLS(&tls.Config{InsecureSkipVerify: true})
		check(err, "could not connect via tls")
	}

	err = ldapConn.Bind(config.Ldap.User, config.Ldap.Password)
	check(err, "could not bind using ldap lookup user")

	searchRequest := ldap.NewSearchRequest(
		config.Ldap.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		strings.Replace(config.Ldap.Filter, "{{username}}", username, -1),
		[]string{"dn"},
		nil,
	)

	sr, err := ldapConn.Search(searchRequest)
	check(err, "error searching a user")

	if len(sr.Entries) > 1 {
		log.Printf("[critical] more than 1 ldap entry found for %v. denying access!\n", username)
	}

	if len(sr.Entries) != 1 {
		return false
	}

	userdn := sr.Entries[0].DN

	err = ldapConn.Bind(userdn, password)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
