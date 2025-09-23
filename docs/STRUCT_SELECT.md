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
