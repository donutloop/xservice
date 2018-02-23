// Copyright 2018 XService, All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

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

	return fmt.Sprintf(CallerTpl, fnc.GetName(), ValueList(parameters)), nil
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

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefCallTpl, identifierList(vars), fnc.GetName(), ValueList(params)))
	return nil
}

func (gen *GoBlockGenerator) DefAssginCall(vars []string, fnc TypeReference, params []string) error {

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

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefAssginCallTpl, identifierList(vars), fnc.GetName(), ValueList(params)))
	return nil
}

func (gen *GoBlockGenerator) DefAppend(vars string, params []string) error {

	if len(vars) == 0 {
		return NewGeneratorErrorString(gen, "values are missing")
	}

	if err := ValidateIdent(vars); err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefAppendTpl, Identifier(vars), ValueList(params)))
	return nil
}

func (gen *GoBlockGenerator) DefMake(vars string, params []string) error {

	if len(vars) == 0 {
		return NewGeneratorErrorString(gen, "values are missing")
	}

	if err := ValidateIdent(vars); err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefCallTpl, Identifier(vars), "make", ValueList(params)))
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

	return fmt.Sprintf(DefSCallTpl, identifierList(vars), fnc, ValueList(params)), nil
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

	return fmt.Sprintf(CallerTpl, fnc.GetName(), ValueList(params)), nil
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

func (gen *GoBlockGenerator) DefLongVar(varName, varType string) error {
	return gen.defVar(varName, varType, true)
}

func (gen *GoBlockGenerator) DefShortVar(varName, varType string) error {
	return gen.defVar(varName, varType, false)
}

func (gen *GoBlockGenerator) defVar(varName, varType string, long bool) error {

	var format string
	if long {
		format = DefVarLongTpl
	} else {
		format = DefVarShortTpl
	}

	if varName == "" {
		return NewGeneratorErrorString(gen, "var is missing")
	}

	if err := ValidateIdent(varName); err != nil {
		return NewGeneratorError(gen, err)
	}

	if varType == "" {
		return NewGeneratorErrorString(gen, "type of var is missing")
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(format, Identifier(varName), varType))
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

func (gen *GoBlockGenerator) DefAssert(vars []string, varName string, typ TypeReference) error {

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

	if typ.GetName() == "" {
		return NewGeneratorErrorString(gen, "type is missing")
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefAssertTpl, identifierList(vars), Identifier(varName), typ.GetName()))
	return nil
}

func (gen *GoBlockGenerator) Command(cmd string, params []string) {
	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(CommandTpl, cmd, identifierList(params)))
}

func (gen *GoBlockGenerator) Return(params ...[]string) {

	var s string
	if params != nil && len(params) == 1 {
		s = fmt.Sprintf(DefReturnWithValuesTpl, identifierList(params[0]))
	} else {
		s = DefReturnTpl
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, s)
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

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(StructAssignTpl, Identifier(typ), Identifier(propertyOfTyp), valueToAssign))
	return nil
}

func (gen *GoBlockGenerator) DefIfWithOwnScopeBegin(caller string, rightSide string, operation token.Token, leftSide string) error {
	if err := gen.validateIf(rightSide, operation, leftSide); err != nil {
		return err
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(IfStatmentWithOwnScopeTpl, caller, rightSide, operation.String(), leftSide))
	return nil
}

func (gen *GoBlockGenerator) DefIfBegin(rightSide string, operation token.Token, leftSide string) error {
	if err := gen.validateIf(rightSide, operation, leftSide); err != nil {
		return err
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(IfStatmentTpl, rightSide, operation.String(), leftSide))
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
	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DefReturnWithValuesTpl, line))
	return nil
}

func (gen *GoBlockGenerator) Defer(fnc TypeReference, parameters []string) error {

	line, err := gen.caller(fnc, parameters)
	if err != nil {
		return err
	}
	gen.MetaData.Lines = append(gen.MetaData.Lines, fmt.Sprintf(DeferTpl, line))
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
	gen.MetaData.Lines = append(gen.MetaData.Lines, code+"\n")
	return nil
}

func (gen *GoBlockGenerator) SliceLiteral(sliceGenerator SliceLiteralGenerator) error {
	code, err := sliceGenerator.Render()
	if err != nil {
		return err
	}
	gen.MetaData.Lines = append(gen.MetaData.Lines, code+"\n")
	return nil
}

func (gen *GoBlockGenerator) InitStruct(statement string, initStructGenerator *InitStructGenerator, pointerReference bool) error {
	s, err := initStructGenerator.Render()
	if err != nil {
		return err
	}

	var format string
	if pointerReference {
		format = fmt.Sprintf("%s &%s \n", statement, s)
	} else {
		format = fmt.Sprintf("%s %s \n", statement, s)
	}

	gen.MetaData.Lines = append(gen.MetaData.Lines, format)
	return nil
}
