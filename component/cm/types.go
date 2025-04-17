package cm

import bcm "go.bytecodealliance.org/cm"

type (
	AnyInteger                                                                        bcm.AnyInteger
	AnyList[T any]                                                                    bcm.AnyList[T]
	AnyResult[Shape, Ok, Err any]                                                     bcm.AnyResult[Shape, Ok, Err]
	AnyVariant[Tag Discriminant, Shape, Align any]                                    bcm.AnyVariant[Tag, Shape, Align]
	BoolResult                                                                        bcm.BoolResult
	Discriminant                                                                      bcm.Discriminant
	HostLayout                                                                        bcm.HostLayout
	List[T any]                                                                       bcm.List[T]
	Option[T any]                                                                     bcm.Option[T]
	Rep                                                                               bcm.Rep
	Resource                                                                          bcm.Resource
	Result[Shape, Ok, Err any]                                                        bcm.Result[Shape, Ok, Err]
	Tuple[T0, T1 any]                                                                 bcm.Tuple[T0, T1]
	Tuple3[T0, T1, T2 any]                                                            bcm.Tuple3[T0, T1, T2]
	Tuple4[T0, T1, T2, T3 any]                                                        bcm.Tuple4[T0, T1, T2, T3]
	Tuple5[T0, T1, T2, T3, T4 any]                                                    bcm.Tuple5[T0, T1, T2, T3, T4]
	Tuple6[T0, T1, T2, T3, T4, T5 any]                                                bcm.Tuple6[T0, T1, T2, T3, T4, T5]
	Tuple7[T0, T1, T2, T3, T4, T5, T6 any]                                            bcm.Tuple7[T0, T1, T2, T3, T4, T5, T6]
	Tuple8[T0, T1, T2, T3, T4, T5, T6, T7 any]                                        bcm.Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]
	Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8 any]                                    bcm.Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]
	Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9 any]                               bcm.Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]
	Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any]                          bcm.Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]
	Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any]                     bcm.Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]
	Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any]                bcm.Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]
	Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any]           bcm.Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]
	Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any]      bcm.Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]
	Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any] bcm.Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]
	Variant[Tag Discriminant, Shape, Align any]                                       bcm.Variant[Tag, Shape, Align]
)
