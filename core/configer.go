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

package core

import (
	"container/list"
	"errors"
	"reflect"

	"github.com/go-spring/spring-core/bean"
)

// Configer 配置函数，不立即执行
type Configer struct {
	bean.Runnable
	name   string
	cond   bean.Condition // 判断条件
	before []string       // 位于哪些配置函数之前
	after  []string       // 位于哪些配置函数之后
}

// newConfiger Configer 的构造函数，fn 不能返回 error 以外的其他值
func newConfiger(name string, fn interface{}, tags []string) *Configer {

	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic(errors.New("fn must be a func"))
	}

	return &Configer{
		name: name,
		Runnable: bean.Runnable{
			Fn:        fn,
			StringArg: bean.NewFnStringBindingArg(fnType, false, tags),
		},
	}
}

// Options 设置 Option 模式函数的参数绑定
func (c *Configer) Options(options ...*bean.OptionArg) *Configer {
	c.OptionArg = &bean.FnOptionBindingArg{Options: options}
	return c
}

// WithCondition 为 Configer 设置一个 Condition
func (c *Configer) WithCondition(cond bean.Condition) *Configer {
	c.cond = cond
	return c
}

// Before 设置当前 Configer 在某些 Configer 之前执行
func (c *Configer) Before(configers ...string) *Configer {
	c.before = append(c.before, configers...)
	return c
}

// After 设置当前 Configer 在某些 Configer 之后执行
func (c *Configer) After(configers ...string) *Configer {
	c.after = append(c.after, configers...)
	return c
}

// getBeforeConfigers 获取当前 Configer 依赖的 Configer 列表
func getBeforeConfigers(configers *list.List, i interface{}) *list.List {
	result := list.New()
	current := i.(*Configer)
	for e := configers.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Configer)

		// 检查是否在当前 Configer 的前面
		for _, name := range c.before {
			if current.name == name {
				result.PushBack(c)
			}
		}

		// 检查是否在当前 Configer 的前面
		for _, name := range current.after {
			if c.name == name {
				result.PushBack(c)
			}
		}
	}
	return result
}
