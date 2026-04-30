/* Minimal LLVM-C declarations for bootstrapping (subset of llvm-c/Core.h).
 * Link with libLLVM (e.g. -lLLVM-C on Unix, LLVM-C.lib on Windows with LLVM in LIB path).
 */
#ifndef KODAE_LLVM_SHIM_H
#define KODAE_LLVM_SHIM_H

#ifdef __cplusplus
extern "C" {
#endif

typedef struct LLVMOpaqueContext LLVMContextRef;
typedef struct LLVMOpaqueModule LLVMModuleRef;
typedef struct LLVMOpaqueBuilder LLVMBuilderRef;
typedef struct LLVMOpaqueType LLVMTypeRef;
typedef struct LLVMOpaqueValue LLVMValueRef;

LLVMContextRef LLVMContextCreate(void);
void LLVMContextDispose(LLVMContextRef C);

LLVMModuleRef LLVMModuleCreateWithNameInContext(const char *ModuleID, LLVMContextRef C);
void LLVMDisposeModule(LLVMModuleRef M);

void LLVMInitializeNativeTarget(void);
void LLVMInitializeNativeAsmPrinter(void);

LLVMTypeRef LLVMInt64TypeInContext(LLVMContextRef C);
LLVMValueRef LLVMConstInt(LLVMTypeRef IntTy, unsigned long long N, int SignExtend);

LLVMTypeRef LLVMFunctionType(LLVMTypeRef ReturnType, LLVMTypeRef *ParamTypes,
                             unsigned ParamCount, int IsVarArg);
LLVMValueRef LLVMAddFunction(LLVMModuleRef M, const char *Name, LLVMTypeRef FunctionTy);

LLVMBuilderRef LLVMCreateBuilderInContext(LLVMContextRef C);
void LLVMDisposeBuilder(LLVMBuilderRef B);

LLVMValueRef LLVMAppendBasicBlockInContext(LLVMContextRef C, LLVMValueRef Fn,
                                            const char *Name);
void LLVMPositionBuilderAtEnd(LLVMBuilderRef B, LLVMValueRef Block);

LLVMValueRef LLVMGetParam(LLVMValueRef Fn, unsigned Index);
LLVMValueRef LLVMBuildAdd(LLVMBuilderRef B, LLVMValueRef LHS, LLVMValueRef RHS,
                          const char *Name);
LLVMValueRef LLVMBuildRet(LLVMBuilderRef B, LLVMValueRef V);

#ifdef __cplusplus
}
#endif

#endif /* KODAE_LLVM_SHIM_H */
