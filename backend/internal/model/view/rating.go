package view

type Rating struct {
	Whole int
	Half  bool
}

func calculateRating(reviews []Review) Rating {
	var rating = 0.0

	if len(reviews) > 0 {
		for _, review := range reviews {
			rating += float64(review.Review.Grade)
		}
		rating = rating / float64(len(reviews))
	}

	half := false
	whole := int(rating)
	if rating-float64(whole) > 0.75 {
		whole += 1
	} else if rating-float64(whole) > 0.25 {
		half = true
	}

	return Rating{
		Whole: whole,
		Half:  half,
	}
}
