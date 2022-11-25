package example

type Topping string

func (t Topping) Valid() bool {
	return toppingSliceContains(toppings, t)
}

var toppings = []Topping{"ketchup", "bacon", "salami", "origaon", "mushrooms", "onions", "olives", "mozzarella"}

func FilterInvalidToping(tt []Topping) []Topping {
	var resp []Topping

	for _, t := range tt {
		if !t.Valid() {
			resp = append(resp, t)
		}
	}

	return resp
}

func toppingSliceContains(tt []Topping, search Topping) bool {
	for _, t := range tt {
		if t == search {
			return true
		}
	}
	return false
}
