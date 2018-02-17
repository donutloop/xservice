package types

import (
	"fmt"
	"go/token"
)

type GoBlockMetaData struct {
	Lines []string
}

type GoBlockGenerator struct {
	MetaData GoBlockMetaData
}

func (gen *GoBlockGenerator) caller(fnc TypeReference, parameters []string) (string, error) {

	if fnc == nil {
		return "", NewGeneratorErrorString(gen, "TypeReference is missing")
	}

	if fnc.GetName() == "" {
		return "", NewGeneratorErrorString(gen, "function is missing")
	}

	if err := ValidateIdent(fnc.GetName()); err != nil {
		return "", NewGeneratorError(gen, err)
	}

	return fmt.Sprintf(CallerTpl, fnc.GetName(), identifierList(parameters)), nil
}

func (gen *GoBlockGenerator) Caller(fnc TypeReference, parameters []string) error {
	line, err := gen.caller(fnc, parameters)
	if err != nil {
		return err
	}
	gen.MetaData.Lines = append(gen.MetaData.Lines, line)
	return nil
}

func (gen *GoBlockGenerator) DefCall(vars []string, fnc TypeReference, params []string) error {

	if fnc == nil {
		return NewGeneratorErrorString(gen, "TypeReference is missing")
	}

	if fnc.GetName() == "" {
		return NewGeneratorErrorString(gen, "function is missing")
	}

	if len(vars) == 0 {
		return NewGeneratorErrorString(gen, "return values are missing")
	}

	if err := ValidateIdents(vars); err != nil {
		return NewGeneratorError(gen, err)
	}

	if err := ValidateIdent(fnc.GetName()); err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefCallTpl, identifierList(vars), fnc.GetName(), identifierList(params)))
	return nil
}

func (gen *GoBlockGenerator) DefAppend(vars string, params []string) error {

	if len(vars) == 0 {
		return NewGeneratorErrorString(gen, "values are missing")
	}

	if err := ValidateIdent(vars); err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefAppendTpl, Identifier(vars), UnsafeIdentifierList(params)))
	return nil
}

func (gen *GoBlockGenerator) DefMake(vars string, params []string) error {

	if len(vars) == 0 {
		return NewGeneratorErrorString(gen, "values are missing")
	}

	if err := ValidateIdent(vars); err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefCallTpl, Identifier(vars), "make", UnsafeIdentifierList(params)))
	return nil
}

func (gen *GoBlockGenerator) SCallWithDefVar(vars []string, fnc string, params []string) (string, error) {

	if fnc == "" {
		return "", NewGeneratorErrorString(gen, "function is missing")
	}

	if err := ValidateIdent(fnc); err != nil {
		return "", NewGeneratorError(gen, err)
	}

	if len(vars) == 0 {
		return "", NewGeneratorErrorString(gen, "return values are missing")
	}

	if err := ValidateIdents(vars); err != nil {
		return "", NewGeneratorError(gen, err)
	}

	return fmt.Sprintf(DefSCallTpl, identifierList(vars), fnc, identifierList(params)), nil
}

func (gen *GoBlockGenerator) SCall(fnc TypeReference, params []string) (string, error) {

	if fnc == nil {
		return "", NewGeneratorErrorString(gen, "TypeReference is missing")
	}

	if fnc.GetName() == "" {
		return "", NewGeneratorErrorString(gen, "function is missing")
	}

	if err := ValidateIdent(fnc.GetName()); err != nil {
		return "", NewGeneratorError(gen, err)
	}

	return fmt.Sprintf(CallerTpl, fnc.GetName(), identifierList(params)), nil
}

func (gen *GoBlockGenerator) DefNew(varName string, typ TypeReference) error {

	if varName == "" {
		return NewGeneratorErrorString(gen, "var is missing")
	}

	if err := ValidateIdent(varName); err != nil {
		return NewGeneratorError(gen, err)
	}

	if typ == nil {
		return NewGeneratorErrorString(gen, "TypeReference is missing")
	}

	if typ.GetName() == "" {
		return NewGeneratorErrorString(gen, "typ is missing")
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefNewTpl, Identifier(varName), typ.GetName()))
	return nil
}

func (gen *GoBlockGenerator) DefVar(varName, varValue string) error {

	if varName == "" {
		return NewGeneratorErrorString(gen, "var is missing")
	}

	if err := ValidateIdent(varName); err != nil {
		return NewGeneratorError(gen, err)
	}

	if varValue == "" {
		return NewGeneratorErrorString(gen, "value of var is missing")
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefVarShortTpl, Identifier(varName), varValue))
	return nil
}

func (gen *GoBlockGenerator) DefStruct(varName, typ, propertyOfTyp string) error {

	if varName == "" {
		return NewGeneratorErrorString(gen, "var is missing")
	}

	if err := ValidateIdent(varName); err != nil {
		return NewGeneratorError(gen, err)
	}

	if typ == "" {
		return NewGeneratorErrorString(gen, "type is missing")
	}

	if propertyOfTyp == "" {
		return NewGeneratorErrorString(gen, "property of typ is missing")
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefStructTpl, Identifier(varName), Identifier(typ), Identifier(propertyOfTyp)))
	return nil
}

func (gen *GoBlockGenerator) DefOperation(varName string, rightSide string, operation token.Token, leftSide string) error {

	if rightSide == "" {
		return NewGeneratorErrorString(gen, "right side of operation is missing")
	}

	if leftSide == "" {
		return NewGeneratorErrorString(gen, "left side of operation is missing")
	}

	if err := ValidateIdent(varName); err != nil {
		return NewGeneratorError(gen, err)
	}

	if err := ValidateOperation(operation); err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefOperationTpl, Identifier(varName), Identifier(rightSide), operation.String(), Identifier(leftSide)))
	return nil
}

func (gen *GoBlockGenerator) DefAssert(vars []string, varName, typ string) error {

	if len(vars) == 0 {
		return NewGeneratorErrorString(gen, "vars is missing")
	}

	if err := ValidateIdents(vars); err != nil {
		return NewGeneratorError(gen, err)
	}

	if varName == "" {
		return NewGeneratorErrorString(gen, "var is missing")
	}

	if err := ValidateIdent(varName); err != nil {
		return NewGeneratorError(gen, err)
	}

	if typ == "" {
		return NewGeneratorErrorString(gen, "type is missing")
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefAssertTpl, identifierList(vars), Identifier(varName), typ))
	return nil
}

func (gen *GoBlockGenerator) Command(cmd string, params []string) {
	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(CommandTpl, cmd, identifierList(params)))
}

func (gen *GoBlockGenerator) Return(params []string) {
	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefReturn, identifierList(params)))
}

func (gen *GoBlockGenerator) StructAssignment(typ, propertyOfTyp string, valueToAssign string) error {

	if typ == "" {
		return NewGeneratorErrorString(gen, "type is missing")
	}

	if propertyOfTyp == "" {
		return NewGeneratorErrorString(gen, "property of typ is missing")
	}

	if valueToAssign == "" {
		return NewGeneratorErrorString(gen, "value is missing")
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(StructAssignTpl, Identifier(typ), Identifier(propertyOfTyp), Identifier(valueToAssign)))
	return nil
}

func (gen *GoBlockGenerator) DefIfWithOwnScopeBegin(caller string, rightSide string, operation token.Token, leftSide string) error {
	if err := gen.validateIf(rightSide, operation, leftSide); err != nil {
		return err
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(IfStatmentWithOwnScopeTpl, caller, Identifier(rightSide), operation.String(), Identifier(leftSide)))
	return nil
}

func (gen *GoBlockGenerator) DefIfBegin(rightSide string, operation token.Token, leftSide string) error {
	if err := gen.validateIf(rightSide, operation, leftSide); err != nil {
		return err
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(IfStatmentTpl, Identifier(rightSide), operation.String(), Identifier(leftSide)))
	return nil
}

func (gen *GoBlockGenerator) validateIf(rightSide string, operation token.Token, leftSide string) error {
	if rightSide == "" {
		return NewGeneratorErrorString(gen, "right side of if is missing")
	}

	if leftSide == "" {
		return NewGeneratorErrorString(gen, "left side of if is missing")
	}

	if err := ValidateIdent(rightSide); err != nil {
		return NewGeneratorError(gen, err)
	}

	if err := ValidateIdent(leftSide); err != nil {
		return NewGeneratorError(gen, err)
	}

	if err := ValidateIfOperation(operation); err != nil {
		return NewGeneratorError(gen, err)
	}
	return nil
}

func (gen *GoBlockGenerator) DefRangeBegin(index string, value string, list string) error {
	if index == "" {
		return NewGeneratorErrorString(gen, "index of range is missing")
	}

	if value == "" {
		return NewGeneratorErrorString(gen, "value of range is missing")
	}

	if err := ValidateIdent(index); err != nil {
		return NewGeneratorError(gen, err)
	}

	if err := ValidateIdent(value); err != nil {
		return NewGeneratorError(gen, err)
	}

	if err := ValidateIdent(list); err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(RangeTpl, Identifier(index), value, Identifier(list)))
	return nil
}

func (gen *GoBlockGenerator) ReturnCaller(fnc TypeReference, parameters []string) error {
	line, err := gen.caller(fnc, parameters)
	if err != nil {
		return err
	}
	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefReturn, line))
	return nil
}

func (gen *GoBlockGenerator) Else() {
	gen.MetaData.Lines = append(gen.MetaData.Lines, ElseStatmentTpl)
}

func (gen *GoBlockGenerator) CloseIf() {
	gen.MetaData.Lines = append(gen.MetaData.Lines, IfEndTpl)
}

func (gen *GoBlockGenerator) CloseRange() {
	gen.MetaData.Lines = append(gen.MetaData.Lines, RangeEndTpl)
}

func (gen *GoBlockGenerator) AnonymousGoFunc(generator *AnonymousFuncGenerator) error {
	code, err := generator.Render()
	if err != nil {
		return err
	}
	gen.MetaData.Lines = append(gen.MetaData.Lines, code+"\n")
	return nil
}

func (gen *GoBlockGenerator) TypeSwitch(switchGenerator SwitchGenerator) error {
	code, err := switchGenerator.Render()
	if err != nil {
		return err
	}
	gen.MetaData.Lines = append(gen.MetaData.Lines, code)
	return nil
}
