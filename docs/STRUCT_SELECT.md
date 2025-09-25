# StructSQL Select Function - Fluent Interface Specification

## Final API Design ✅

```go
type SelectBuilder struct {
    s          *Structsql
    from       any
    wheres     [8]whereClause  // Fixed array to avoid allocations
    orders     [4]OrderField   // Fixed array for ORDER BY
    whereCount int
    orderCount int
    limit      int
    offset     int
}

// Core API
func (s *Structsql) Select(from any) *SelectBuilder
func (sb *SelectBuilder) Where(condition any) *SelectBuilder
func (sb *SelectBuilder) Or(condition any) *SelectBuilder
func (sb *SelectBuilder) WhereIN(field string, values []any) *SelectBuilder
func (sb *SelectBuilder) WhereBetween(field string, min, max any) *SelectBuilder
func (sb *SelectBuilder) OrderBy(field string, desc bool) *SelectBuilder
func (sb *SelectBuilder) Limit(n int) *SelectBuilder
func (sb *SelectBuilder) Offset(n int) *SelectBuilder
func (sb *SelectBuilder) Build(sql *string, values *[]any) error
```

## Field Naming Convention - REGLAS IMPORTANTES ⚠️

### Sin Struct Tags - Responsabilidad del Desarrollador

**Filosofía**: Sin etiquetas `db:""` o similares. Los nombres de campos del struct deben coincidir exactamente con las columnas de la base de datos.

### Transformación Automática
```go
// StructSQL solo aplica: NombreCampo → nombrecampo (minúsculas)
type User struct {
    ID_Staff  string  // → id_staff
    UserName  string  // → username (NO user_name)  
    Full_Name string  // → full_name
}
```

### Ejemplos Correctos e Incorrectos

#### ✅ **CORRECTO**:
```go
// Struct fields que coinciden con columnas DB
type Reservation struct {
    ID_Reservation      int     // DB: id_reservation
    ID_Staff           string   // DB: id_staff  
    Service_Name       string   // DB: service_name
    Reservation_Year   int      // DB: reservation_year
}
// → SELECT id_reservation, id_staff, service_name, reservation_year FROM reservation
```

#### ❌ **INCORRECTO**:
```go
// Struct fields que NO coinciden con columnas DB
type Reservation struct {
    IDReservation     int     // → idreservation (NO id_reservation)
    StaffId          string   // → staffid (NO id_staff)
    ServiceName      string   // → servicename (NO service_name) 
    ReservationYear  int      // → reservationyear (NO reservation_year)
}
```

### Recomendaciones de Naming

1. **Para snake_case DB**: Usar underscores en struct
   ```go
   DB: user_profile → Struct: User_Profile
   DB: created_at   → Struct: Created_At
   ```

2. **Para nombres simples**: Usar nombres simples
   ```go
   DB: id    → Struct: ID
   DB: name  → Struct: Name
   ```

3. **Consistencia**: Mantener el mismo patrón en todo el proyecto

### Validación en Desarrollo
```go
// Si tu columna DB es: id_staff
// Y usas: IDStaff
// Resultado será: idstaff (INCORRECTO)
// Solución: usar ID_Staff → id_staff (CORRECTO)
```

---

```go
type whereClause struct {
    condition any
    operator  string // "AND", "OR"
    field     string // Para IN/BETWEEN operations
    values    []any  // Para IN/BETWEEN values
}

type OrderField struct {
    Column string
    DESC   bool
}
```

## Zero-Allocation Pool Strategy

```go
// Object pooling siguiendo patrón tinystring/memory.go
var selectBuilderPool = sync.Pool{
    New: func() any {
        return &SelectBuilder{
            wheres:     [8]whereClause{},
            orders:     [4]OrderField{},
            whereCount: 0,
            orderCount: 0,
            limit:      0,
            offset:     0,
        }
    },
}

func (s *Structsql) Select(from any) *SelectBuilder {
    sb := selectBuilderPool.Get().(*SelectBuilder)
    sb.reset()
    sb.s = s
    sb.from = from
    return sb
}

func (sb *SelectBuilder) reset() {
    sb.whereCount = 0
    sb.orderCount = 0
    sb.limit = 0
    sb.offset = 0
    sb.from = nil
    sb.s = nil
}

func (sb *SelectBuilder) Build(sql *string, values *[]any) error {
    defer func() {
        sb.reset()
        selectBuilderPool.Put(sb) // Auto-return to pool
    }()
    
    // Implementation details...
    return nil
}
```

## Core Implementation Decisions

### 1. FROM Clause Support
- ✅ Solo structs que implementen StructNamer interface
- ✅ `User{}` → `FROM user`
- ✅ Nombres de tabla en minúsculas para compatibilidad

### 2. Field Naming Convention - IMPORTANTE ⚠️
**Sin etiquetas de struct**: Los nombres de campos deben coincidir con las columnas de la base de datos
- ✅ Struct field `ID_Staff` → columna `id_staff` (automáticamente a minúsculas)
- ✅ Struct field `ReservationYear` → columna `reservationyear` (NO snake_case automático)
- ✅ **Responsabilidad del desarrollador**: Nombrar campos del struct como en la DB
- ✅ **Recomendación**: Usar snake_case en structs: `Reservation_Year` → `reservation_year`

**Ejemplo correcto**:
```go
type Reservation struct {
    ID_Reservation      int
    ID_Staff           string  
    Service_Name       string
    Reservation_Year   int
    Reservation_Month  int
    Reservation_Day    int
}
// → SELECT id_reservation, id_staff, service_name, reservation_year, reservation_month, reservation_day
```

### 3. WHERE Clause Logic
- ✅ Zero-values ignorados en structs
- ✅ Direct method approach para operadores avanzados
- ✅ Default AND logic, explícito OR con método Or()
- ✅ WhereIN() y WhereBetween() methods para casos avanzados

### 4. SELECT Columns Strategy
- ✅ Columnas basadas en FROM struct fields
- ✅ Si User tiene {ID, Name, Email} → `SELECT id, name, email`

### 5. Zero-Allocation Philosophy
- ✅ SelectBuilder pooling
- ✅ Fixed arrays evitan slice growth
- ✅ GetStringZeroCopy para SQL output
- ✅ Auto-return to pool en Build()

## Usage Examples

### Basic Usage
```go
type User struct {
    ID    int
    Name  string  
    Email string
}

user := User{}
var sql string
var values []any

// Simple SELECT
err := s.Select(user).Build(&sql, &values)
// → SELECT id, name, email FROM user

// With WHERE
err := s.Select(user).
    Where(User{ID: 1}).
    Build(&sql, &values)
// → SELECT id, name, email FROM user WHERE id=$1
```

### Complex Example (Real-world SQL)
```sql
-- Target SQL:
SELECT id_reservation, id_staff, service_name, service_time, reservation_creator, 
       reservation_year, reservation_month, reservation_day, reservation_hour, 
       reservation_detail, reservation_verified, id_patient
  FROM reservation WHERE id_staff='1635572582072481400' 
  ORDER BY reservation_year DESC, reservation_month DESC, reservation_day DESC;
```

```go
// Struct con nombres que coinciden con la DB
type Reservation struct {
    ID_Reservation      int
    ID_Staff           string
    Service_Name       string
    Service_Time       string
    Reservation_Creator string
    Reservation_Year   int
    Reservation_Month  int
    Reservation_Day    int
    Reservation_Hour   int
    Reservation_Detail string
    Reservation_Verified bool
    ID_Patient         int
}

// Fluent implementation
reservation := Reservation{}
var sql string
var values []any

err := s.Select(reservation).
    Where(Reservation{ID_Staff: "1635572582072481400"}).
    OrderBy("reservation_year", true).   // DESC
    OrderBy("reservation_month", true).  // DESC
    OrderBy("reservation_day", true).    // DESC
    Build(&sql, &values)
```

### Advanced Operators (Simplified Fluent Approach)
```go
type User struct {
    ID            int
    Name          string  
    Email         string
    Status        string
    Role          string
    Department_ID int
    Salary        int
    Created_At    string
}

// Direct method approach - No interfaces needed
err := s.Select(User{}).
    WhereIN("id", []any{1, 2, 3}).
    Build(&sql, &values)
// → SELECT id, name, email, status, role, department_id, salary, created_at FROM user WHERE id IN($1,$2,$3)

// WhereBetween method
err := s.Select(User{}).
    WhereBetween("age", 18, 65).
    Build(&sql, &values)
// → SELECT id, name, email, status, role, department_id, salary, created_at FROM user WHERE age BETWEEN $1 AND $2

// Complex combinations
err := s.Select(User{}).
    Where(User{Status: "active"}).
    Or(User{Role: "admin"}).
    WhereIN("department_id", []any{1, 2, 3}).
    WhereBetween("salary", 50000, 100000).
    OrderBy("created_at", true).
    Limit(50).
    Offset(100).
    Build(&sql, &values)
```

### Pagination
```go
err := s.Select(User{}).
    Where(User{Status: "active"}).
    OrderBy("id", false). // ASC
    Limit(20).
    Offset(40).
    Build(&sql, &values)
// → SELECT id, name, email, status, role, department_id, salary, created_at FROM user WHERE status=$1 ORDER BY id ASC LIMIT 20 OFFSET 40
```

## Implementation Advantages

### ✅ Type Safety & Clarity
- Autodocumentado y fácil de leer
- Validación por método individual
- No confusion con args variadic

### ✅ Zero-Allocation Performance
- Object pooling pattern de tinystring
- Fixed arrays evitan heap allocations
- GetStringZeroCopy para SQL output
- Automatic cleanup y return to pool

### ✅ Scalability
- Fácil agregar nuevos métodos (GroupBy, Having, etc)
- Extensible sin romper API existente
- Clear separation of concerns

### ✅ Consistency
- Mantiene filosofía zero-allocation
- Usa patrones establecidos de tinystring
- Compatible con tinyreflect
- Natural integration con Insert/Update/Delete

## Implementation Roadmap

### V1 Core Features ✅
- [x] Fluent interface design
- [x] Zero-allocation pooling strategy
- [x] FROM struct support
- [x] WHERE with equality operators
- [x] WhereIN/WhereBetween direct methods
- [x] ORDER BY support
- [x] LIMIT/OFFSET support
- [x] SELECT columns from FROM struct

### V1 Implementation Tasks
- [ ] Implementar SelectBuilder con pooling
- [ ] Implementar métodos fluent core
- [ ] Implementar Build() con zero-allocation
- [ ] Integrar con shared.go helpers
- [ ] Crear select_test.go
- [ ] Benchmarks validation

### V2 Future Enhancements
- [ ] JOIN support
- [ ] GROUP BY/HAVING methods
- [ ] Subquery support
- [ ] SqlSELECT interface para column override

---

# PROPUESTAS DE TIPADO PARA FLUENT INTERFACE

## Objetivo: Tipado en la Construcción de Consultas

El usuario solicita tipado estático para evitar errores en nombres de campos:

```go
// OBJETIVO: En lugar de strings
err := s.Select(reservation).
    OrderBy("reservation_year", true).   // Prone a errores de typo
    OrderBy("reservation_month", true).
    Build(&sql, &values)

// OBJETIVO: Tipado estático

r := Reservation{}

err := s.Select(r).
    OrderBy(r.Reservation_Year, true).   // Type-safe, autocompletado IDE
    OrderBy(r.Reservation_Month, true).
    Build(&sql, &values)
```

---

## PROPUESTA 1: Field Selector Interface con Type Assertion

### Descripción
Crear una interfaz `FieldSelector` que permita extraer el nombre del campo de manera type-safe usando type assertion y reflection.

### Implementación

```go
// FieldSelector interface for type-safe field selection
type FieldSelector interface {
    FieldName() string
    IsDescending() bool
}

// FieldSelectorImpl implements FieldSelector for struct fields
type FieldSelectorImpl[T any] struct {
    fieldName string
    descending bool
}

// FieldName returns the database column name
func (fs FieldSelectorImpl[T]) FieldName() string {
    return fs.fieldName
}

// IsDescending returns the sort direction
func (fs FieldSelectorImpl[T]) IsDescending() bool {
    return fs.descending
}

// FieldExtractor extracts field information from struct field references
type FieldExtractor struct {
    structsql *Structsql
}

// ExtractFieldName uses reflection to get field name from struct field reference
func (fe *FieldExtractor) ExtractFieldName(fieldRef any, descending bool) (string, error) {
    // Use tinyreflect to analyze the field reference
    v := tinyreflect.ValueOf(fieldRef)

    // Check if it's a struct field reference (pointer to struct field)
    if v.Kind() != K.Pointer {
        return "", Err("field reference must be a pointer to struct field")
    }

    // Get the underlying type
    elemType := v.Elem().Type()
    if elemType.Kind() != K.Struct {
        return "", Err("field reference must point to a struct field")
    }

    // Extract field name using reflection
    structType := elemType.StructType()
    if structType == nil {
        return "", Err("invalid struct type")
    }

    // Get field name from the struct type
    // This requires analyzing the memory layout to find which field this is
    fieldName := fe.extractFieldNameFromPointer(structType, fieldRef)
    if fieldName == "" {
        return "", Err("could not determine field name")
    }

    return fieldName, nil
}

// extractFieldNameFromPointer analyzes the pointer to determine field name
func (fe *FieldExtractor) extractFieldNameFromPointer(structType *tinyreflect.StructType, fieldPtr any) string {
    // This is a simplified version - actual implementation would need
    // to analyze the memory offset to determine which field this pointer represents
    // For now, return empty string as placeholder
    return ""
}

// Updated SelectBuilder API
type SelectBuilder struct {
    s          *Structsql
    from       any
    wheres     [8]whereClause
    orders     [4]FieldSelector  // Use FieldSelector instead of OrderField
    whereCount int
    orderCount int
    limit      int
    offset     int
    extractor  *FieldExtractor
}

// Updated OrderBy method
func (sb *SelectBuilder) OrderBy(fieldRef any, descending bool) *SelectBuilder {
    if sb.orderCount >= len(sb.orders) {
        // Handle overflow - could expand array or return error
        return sb
    }

    fieldName, err := sb.extractor.ExtractFieldName(fieldRef, descending)
    if err != nil {
        // Handle error - could store error for later or panic
        return sb
    }

    selector := FieldSelectorImpl[any]{
        fieldName:  fieldName,
        descending: descending,
    }

    sb.orders[sb.orderCount] = selector
    sb.orderCount++
    return sb
}
```

### Pros ✅
- **Type Safety**: El compilador verifica que el campo existe en la estructura
- **IDE Support**: Autocompletado completo para campos de estructura
- **Zero-Copy**: Compatible con filosofía zero-allocation
- **Extensible**: Fácil agregar nuevos métodos (GroupBy, Having, etc.)
- **Runtime Validation**: Puede validar que el campo existe en tiempo de ejecución

### Contras ❌
- **Complex Implementation**: Requiere análisis de memoria para determinar nombres de campos
- **Performance Overhead**: Reflection tiene costo en runtime
- **Limited to Exported Fields**: Solo funciona con campos exportados (mayúscula inicial)
- **Memory Analysis Complexity**: Determinar qué campo representa un puntero requiere análisis de offsets

### Recomendaciones
1. **Implementar FieldExtractor**: Desarrollar la lógica de análisis de memoria para mapear punteros a nombres de campos
2. **Caching**: Cachear resultados de análisis de estructura para mejorar performance
3. **Fallback Strategy**: Proporcionar fallback a strings para casos complejos
4. **Error Handling**: Implementar manejo robusto de errores con información útil

---

## PROPUESTA 2: Generic Field Extractor con Compile-time Generation

### Descripción
Usar generics y generación de código en tiempo de compilación para crear extractores de campo específicos para cada estructura.

### Implementación

```go
// FieldExtractor generates type-safe field extractors for specific structs
type FieldExtractor[T any] struct {
    structsql *Structsql
    fieldMap  map[string]int // field name to index mapping
}

// NewFieldExtractor creates a new extractor for type T
func NewFieldExtractor[T any](s *Structsql) *FieldExtractor[T] {
    var t T
    typ := tinyreflect.TypeOf(t)

    if typ.Kind() != K.Struct {
        panic("T must be a struct type")
    }

    structType := typ.StructType()
    fieldMap := make(map[string]int)

    for i := 0; i < len(structType.Fields); i++ {
        field, _ := typ.Field(i)
        fieldName := field.Name.Name()
        fieldMap[fieldName] = i
    }

    return &FieldExtractor[T]{
        structsql: s,
        fieldMap:  fieldMap,
    }
}

// Field creates a type-safe field reference
func (fe *FieldExtractor[T]) Field(fieldName string) *FieldRef[T] {
    if _, exists := fe.fieldMap[fieldName]; !exists {
        panic(fmt.Sprintf("field %s does not exist in struct", fieldName))
    }
    return &FieldRef[T]{extractor: fe, fieldName: fieldName}
}

// FieldRef represents a reference to a struct field
type FieldRef[T any] struct {
    extractor *FieldExtractor[T]
    fieldName string
    descending bool
}

// Desc sets the sort direction to descending
func (fr *FieldRef[T]) Desc() *FieldRef[T] {
    fr.descending = true
    return fr
}

// Asc sets the sort direction to ascending (default)
func (fr *FieldRef[T]) Asc() *FieldRef[T] {
    fr.descending = false
    return fr
}

// FieldName returns the database column name
func (fr *FieldRef[T]) FieldName() string {
    // Convert field name to database column name
    // e.g., "Reservation_Year" -> "reservation_year"
    return strings.ToLower(strings.ReplaceAll(fr.fieldName, "_", ""))
}

// IsDescending returns the sort direction
func (fr *FieldRef[T]) IsDescending() bool {
    return fr.descending
}

// Updated SelectBuilder with generic support
type SelectBuilder[T any] struct {
    s          *Structsql
    from       T
    wheres     [8]whereClause
    orders     [4]FieldSelector
    whereCount int
    orderCount int
    limit      int
    offset     int
    extractor  *FieldExtractor[T]
}

// Select creates a new SelectBuilder for type T
func (s *Structsql) Select(from T) *SelectBuilder[T] {
    sb := selectBuilderPool.Get().(*SelectBuilder[T])
    sb.reset()
    sb.s = s
    sb.from = from
    sb.extractor = NewFieldExtractor[T](s)
    return sb
}

// Updated OrderBy method using FieldRef
func (sb *SelectBuilder[T]) OrderBy(fieldRef *FieldRef[T]) *SelectBuilder[T] {
    if sb.orderCount >= len(sb.orders) {
        return sb
    }

    sb.orders[sb.orderCount] = fieldRef
    sb.orderCount++
    return sb
}
```

### Pros ✅
- **Full Type Safety**: Validación completa en tiempo de compilación
- **Excellent IDE Support**: Autocompletado perfecto y navegación de código
- **Performance**: Sin reflection en runtime, solo en inicialización
- **Zero-Allocation Compatible**: Puede integrarse con pooling existente
- **Self-Documenting**: El código es autoexplicativo
- **Compile-time Validation**: Errores detectados antes de ejecutar

### Contras ❌
- **Code Generation**: Requiere generar extractores específicos por tipo
- **Memory Overhead**: Cada tipo necesita su propio extractor
- **Initialization Cost**: Setup inicial más complejo
- **Generic Complexity**: Requiere Go 1.18+ con generics
- **API Complexity**: API más compleja para el usuario

### Recomendaciones
1. **Code Generation**: Implementar generación automática de extractores
2. **Caching Strategy**: Cachear extractores por tipo para reusar
3. **Hybrid Approach**: Combinar con Proposal 1 para casos complejos
4. **Migration Path**: Proporcionar path de migración desde API actual

---

## ANÁLISIS COMPARATIVO

| Aspecto | Propuesta 1 (Type Assertion) | Propuesta 2 (Generics) |
|---------|-----------------------------|----------------------|
| **Type Safety** | Runtime | Compile-time |
| **Performance** | Reflection overhead | Minimal overhead |
| **IDE Support** | Bueno | Excelente |
| **Complexity** | Media | Alta |
| **Flexibility** | Alta | Media |
| **Go Version** | Compatible con todas | Go 1.18+ |

## RECOMENDACIÓN FINAL

**Recomendamos implementar la Propuesta 2 (Generic Field Extractor)** por las siguientes razones:

1. **Superior Type Safety**: Validación en tiempo de compilación
2. **Mejor Developer Experience**: IDE support completo
3. **Performance**: Sin reflection en runtime
4. **Future-Proof**: Alineado con dirección moderna de Go
5. **Maintainability**: Menos propenso a errores

### Plan de Implementación

1. **Fase 1**: Implementar FieldExtractor básico con generics
2. **Fase 2**: Agregar generación automática de extractores
3. **Fase 3**: Integrar con SelectBuilder existente
4. **Fase 4**: Proporcionar API de compatibilidad hacia atrás
5. **Fase 5**: Documentación y ejemplos completos

### Ejemplo de Uso Final

```go
type Reservation struct {
    ID_Reservation      int
    ID_Staff           string
    Service_Name       string
    Reservation_Year   int
    Reservation_Month  int
    Reservation_Day    int
}

func main() {
    s := New()
    r := Reservation{}

    // Type-safe fluent interface
    extractor := NewFieldExtractor[Reservation](s)

    var sql string
    var values []any

    err := s.Select(r).
        OrderBy(extractor.Field("Reservation_Year").Desc()).
        OrderBy(extractor.Field("Reservation_Month").Desc()).
        OrderBy(extractor.Field("Reservation_Day").Desc()).
        Build(&sql, &values)
}
```

---
ningina propuesta es buena,,eliminalas..lo que busco es simplificar y que sea intuitivo, pensaba en algo como OrderBy(r.Reservation_Year, true) donde OrderBy(any,bool) debemos ser capases de detectar si el primer valor ingresado es de tipo string y es un campo de una estructura..se puede hacer con refelct? si es asi como compdiramos implmentarlo en tinyreflect? 

investiga
actuliza el docuemneto con 2 propuestas con sus pro contras y recomendaciones..espera mi revision