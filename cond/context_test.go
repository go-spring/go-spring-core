/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cond_test

import (
	"errors"
	"testing"

	"github.com/go-spring/spring-core/bean"
	"github.com/go-spring/spring-core/cond"
	"github.com/go-spring/spring-core/core"
	"github.com/go-spring/spring-core/util"
)

type Teacher interface {
	Course() string
}

type historyTeacher struct {
	name string
}

func newHistoryTeacher(name string) *historyTeacher {
	return &historyTeacher{
		name: name,
	}
}

func (t *historyTeacher) Course() string {
	return "history"
}

type Student struct {
	Teacher Teacher
	Room    string
}

// 入参可以进行注入或者属性绑定，返回值可以是 struct、map、slice、func 等。
func NewStudent(teacher Teacher, room string) Student {
	return Student{
		Teacher: teacher,
		Room:    room,
	}
}

// 入参可以进行注入或者属性绑定，返回值可以是 struct、map、slice、func 等。
func NewPtrStudent(teacher Teacher, room string) *Student {
	return &Student{
		Teacher: teacher,
		Room:    room,
	}
}

var defaultClassOption = ClassOption{
	className: "default",
}

type ClassOption struct {
	className string
	students  []*Student
	floor     int
}

type ClassOptionFunc func(opt *ClassOption)

func withClassName(className string, floor int) ClassOptionFunc {
	return func(opt *ClassOption) {
		opt.className = className
		opt.floor = floor
	}
}

func withStudents(students []*Student) ClassOptionFunc {
	return func(opt *ClassOption) {
		opt.students = students
	}
}

type ClassRoom struct {
	President string `value:"${president}"`
	className string
	floor     int
	students  []*Student
	desktop   Desktop
}

type Desktop interface {
}

type MetalDesktop struct {
}

func (cls *ClassRoom) Desktop() Desktop {
	return cls.desktop
}

func NewClassRoom(options ...ClassOptionFunc) ClassRoom {
	opt := defaultClassOption
	for _, fn := range options {
		fn(&opt)
	}
	return ClassRoom{
		className: opt.className,
		students:  opt.students,
		floor:     opt.floor,
		desktop:   &MetalDesktop{},
	}
}

type ServerInterface interface {
	Consumer() *Consumer
	ConsumerT() *Consumer
	ConsumerArg(i int) *Consumer
}

type Server struct {
	Version string `value:"${server.version}"`
}

func NewServerInterface() ServerInterface {
	return new(Server)
}

type Consumer struct {
	s *Server
}

func (s *Server) Consumer() *Consumer {
	if nil == s {
		panic(errors.New("server is nil"))
	}
	return &Consumer{s}
}

func (s *Server) ConsumerT() *Consumer {
	return s.Consumer()
}

func (s *Server) ConsumerArg(i int) *Consumer {
	if nil == s {
		panic(errors.New("server is nil"))
	}
	return &Consumer{s}
}

type Service struct {
	Consumer *Consumer `autowire:""`
}

func TestDefaultSpringContext(t *testing.T) {

	t.Run("bean:test_ctx:", func(t *testing.T) {

		ctx := core.NewApplicationContext()
		ctx.RegisterBean(bean.Ref(&BeanZero{5}).WithCondition(cond.
			OnProfile("test").
			And().
			OnMissingBean("null"),
		))

		ctx.AutoWireBeans()

		var b *BeanZero
		ok := ctx.GetBean(&b)
		util.AssertEqual(t, ok, false)
	})

	t.Run("bean:test_ctx:test", func(t *testing.T) {

		ctx := core.NewApplicationContext()
		ctx.Profile("test")
		ctx.RegisterBean(bean.Ref(&BeanZero{5}).WithCondition(cond.OnProfile("test")))
		ctx.AutoWireBeans()

		var b *BeanZero
		ok := ctx.GetBean(&b)
		util.AssertEqual(t, ok, true)
	})

	t.Run("bean:test_ctx:stable", func(t *testing.T) {

		ctx := core.NewApplicationContext()
		ctx.Profile("stable")
		ctx.RegisterBean(bean.Ref(&BeanZero{5}).WithCondition(cond.OnProfile("test")))
		ctx.AutoWireBeans()

		var b *BeanZero
		ok := ctx.GetBean(&b)
		util.AssertEqual(t, ok, false)
	})

	t.Run("option withClassName Condition", func(t *testing.T) {

		ctx := core.NewApplicationContext()
		ctx.Property("president", "CaiYuanPei")
		ctx.Property("class_floor", 2)
		ctx.RegisterBean(bean.Make(NewClassRoom).Options(
			bean.NewOptionArg(withClassName,
				"${class_name:=二年级03班}",
				"${class_floor:=3}",
			).WithCondition(cond.OnProperty("class_name_enable")),
		))
		ctx.AutoWireBeans()

		var cls *ClassRoom
		ctx.GetBean(&cls)

		util.AssertEqual(t, cls.floor, 0)
		util.AssertEqual(t, len(cls.students), 0)
		util.AssertEqual(t, cls.className, "default")
		util.AssertEqual(t, cls.President, "CaiYuanPei")
	})

	t.Run("option withClassName Apply", func(t *testing.T) {
		c := cond.OnProperty("class_name_enable")

		ctx := core.NewApplicationContext()
		ctx.Property("president", "CaiYuanPei")
		ctx.RegisterBean(bean.Make(NewClassRoom).Options(
			bean.NewOptionArg(withClassName,
				"${class_name:=二年级03班}",
				"${class_floor:=3}",
			).WithCondition(c),
		))
		ctx.AutoWireBeans()

		var cls *ClassRoom
		ctx.GetBean(&cls)

		util.AssertEqual(t, cls.floor, 0)
		util.AssertEqual(t, len(cls.students), 0)
		util.AssertEqual(t, cls.className, "default")
		util.AssertEqual(t, cls.President, "CaiYuanPei")
	})

	t.Run("method bean cond", func(t *testing.T) {

		ctx := core.NewApplicationContext()
		ctx.Property("server.version", "1.0.0")
		parent := ctx.RegisterBean(bean.Ref(new(Server)))
		ctx.RegisterBean(bean.Child(parent, "Consumer").WithCondition(cond.OnProperty("consumer.enable")))
		ctx.AutoWireBeans()

		var s *Server
		ok := ctx.GetBean(&s)
		util.AssertEqual(t, ok, true)
		util.AssertEqual(t, s.Version, "1.0.0")

		var c *Consumer
		ok = ctx.GetBean(&c)
		util.AssertEqual(t, ok, false)
	})

	t.Run("fn method bean cond", func(t *testing.T) {

		ctx := core.NewApplicationContext()
		ctx.Property("server.version", "1.0.0")
		ctx.RegisterBean(bean.Make(NewServerInterface))
		ctx.RegisterBean(bean.MethodFunc(ServerInterface.ConsumerT).WithCondition(cond.OnProperty("consumer.enable")))
		ctx.AutoWireBeans()

		var si ServerInterface
		ok := ctx.GetBean(&si)
		util.AssertEqual(t, ok, true)

		s := si.(*Server)
		util.AssertEqual(t, s.Version, "1.0.0")

		var c *Consumer
		ok = ctx.GetBean(&c)
		util.AssertEqual(t, ok, false)
	})
}

func TestDefaultSpringContext_ParentNotRegister(t *testing.T) {

	ctx := core.NewApplicationContext()
	parent := ctx.RegisterBean(bean.Make(NewServerInterface).
		WithCondition(cond.OnProperty("server.is.nil")))
	ctx.RegisterBean(bean.Child(parent, "Consumer"))

	ctx.AutoWireBeans()

	var s *Server
	ok := ctx.GetBean(&s)
	util.AssertEqual(t, ok, false)

	var c *Consumer
	ok = ctx.GetBean(&c)
	util.AssertEqual(t, ok, false)
}

func TestDefaultSpringContext_ChainConditionOnBean(t *testing.T) {
	for i := 0; i < 20; i++ { // 不要排序
		ctx := core.NewApplicationContext()
		ctx.RegisterBean(bean.Ref(new(string)).WithCondition(cond.OnBean("*bool")))
		ctx.RegisterBean(bean.Ref(new(bool)).WithCondition(cond.OnBean("*int")))
		ctx.RegisterBean(bean.Ref(new(int)).WithCondition(cond.OnBean("*float")))
		ctx.AutoWireBeans()
		util.AssertEqual(t, len(ctx.GetBeanDefinitions()), 0)
	}
}

func TestDefaultSpringContext_ConditionOnBean(t *testing.T) {
	ctx := core.NewApplicationContext()

	c := cond.
		OnMissingProperty("Null").
		Or().
		OnProfile("test")

	ctx.RegisterBean(bean.Ref(&BeanZero{5}).WithCondition(cond.
		On(c).
		And().
		OnMissingBean("null"),
	))

	ctx.RegisterBean(bean.Ref(new(BeanOne)).WithCondition(cond.
		On(c).
		And().
		OnMissingBean("null"),
	))

	ctx.RegisterBean(bean.Ref(new(BeanTwo)).WithCondition(cond.OnBean("*cond_test.BeanOne")))
	ctx.RegisterBean(bean.Ref(new(BeanTwo)).WithName("another_two").WithCondition(cond.OnBean("Null")))

	ctx.AutoWireBeans()

	var two *BeanTwo
	ok := ctx.GetBean(&two, "")
	util.AssertEqual(t, ok, true)

	ok = ctx.GetBean(&two, "another_two")
	util.AssertEqual(t, ok, false)
}

func TestDefaultSpringContext_ConditionOnMissingBean(t *testing.T) {

	for i := 0; i < 20; i++ { // 测试 FindBean 无需绑定，不要排序
		ctx := core.NewApplicationContext()

		ctx.RegisterBean(bean.Ref(&BeanZero{5}))
		ctx.RegisterBean(bean.Ref(new(BeanOne)))

		ctx.RegisterBean(bean.Ref(new(BeanTwo)).WithCondition(cond.OnMissingBean("*cond_test.BeanOne")))
		ctx.RegisterBean(bean.Ref(new(BeanTwo)).WithName("another_two").WithCondition(cond.OnMissingBean("Null")))

		ctx.AutoWireBeans()

		var two *BeanTwo
		ok := ctx.GetBean(&two, "")
		util.AssertEqual(t, ok, true)

		ok = ctx.GetBean(&two, "another_two")
		util.AssertEqual(t, ok, true)
	}
}
