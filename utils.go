package rego

// If 是一个泛型三元运算符模拟函数
func If[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
