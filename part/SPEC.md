# Part

[todo]

## Sequence

A sequence is a range of characters to iterate through, a list of character literals to iterate through, or a list of one or more options.

A `SequenceDefinition` can parse, generate, and validate `SequenceInstances` or strings.

`SequencesDefinitions` are used for part categories and part numbers.

A part category will have a sequence, which has a fk to a sequence definition, and a string value.

A part's fully qualified value will be validated against its parent category's sequence definition.

### SequenceDefinition

- value: range of acceptable characters
- parse(seq: str) -> returns a SequenceInstance if valid or errors / None
- generate(curr: str) -> returns the next value in a sequence based on `curr`
- validate(seq: str) -> if parse(seq): return True else return False

### SequenceInstance

- literal: str
- definition: SequenceDefinition


## Part DB Models

### PartCategory
 - id: pk
 - name: str
 - description: str
 - parent_category: ForeignKey(PartCategory)
 - value: // TODO leverage sequences, should be unique within its parent category
 - created_at: datetime
 - updated_at: datetime

### AbstractPart
 - id: pk
 - name: str
 - description: str
 - category: ForeignKey(PartCategory)
 - part_numer: // TODO leverage sequences, based on this and the category, should be unique within parent category
 - created_at: datetime
 - updated_at: datetime

### PartRevision
 - id: pk
 - part: ForeignKey(AbstractPart)
 - revision_code: str // unique_together(part_id, revision_code), should leverage sequences
 - revision_number: int // Auto-incrementing for each part, managed by DB
 - release_status: ForeignKey(ReleaseStatus)
 - created_at: datetime
 - updated_at: datetime

### TODO PartRevisionConstraint

For specifying parts that are governed by constraints, where the part revision can be by any of a subset of parts that satisfy the constraints. Example: a resistor with a specific footprint, temperature characteristics, COO, etc.

- id: pk
- constraint: TODO
- target: ForeignKey(PartRevision, backref='constraints')

### AlternatePartRevision

For specifying alternates to a given part revision.

- id: pk
- target: ForeignKey(PartRevision)
- alternate: ForeignKey(PartRevision)

### BillOfMaterials
 - id: pk
 - name: str
 - description: str
 - parent_part_revision: ForeignKey(PartRevision)
 - bom_type: ENUM('EBOM', 'MBOM', 'SBOM')
 - created_at: datetime
 - updated_at: datetime

<!-- ## UnitOfMeasure

TODO This seems over wrought, perhaps? But also it moves this distinct class of functionality out of the Parts proper and into its own realm.

 - id: pk
 - name: str
 - description: str
 - short_name: str
 - conversions: ???

## Quantity
 - id: pk
 - unit_of_measure: ForeignKey(UnitOfMeasure)
 - value: Union[int, float, decimal (fixed point)] // should be fixed point, i.e., we don't want to lose information when we store/call/use this data -->

### BillOfMaterialsEntry
 - id: pk
 - bom: ForeignKey(BillOfMaterials, backref='entries')
 - part_revision: ForeignKey(PartRevision)
 - quantity: ForeignKey(Quantity)
 - created_at: datetime
 - updated_at: datetime

### ReferenceDesignator
 - id: pk
 - bom_entry: ForeignKey(BillOfMaterialsEntry, backref='reference_designators')
 - designator: str // Enforces uniqueness per bom_entry

### PartAlias
 - id: pk
 - name: str
 - description: str
 - part_revision: ForeignKey(PartRevision)
 - created_at: datetime
 - updated_at: datetime

### AbstractPartAlias

could be used as external product name that always tracks the latest rev, could also be used to track mfr pn not including rev

 - id: pk
 - name: str
 - description: str
 - abstract_part: ForeignKey(AbstractPart)
 - revision_qualifier: str // e.g., 'latest' or 
 - created_at: datetime
 - updated_at: datetime

### BillOfMaterialsVariant
 - id: pk
 - name: str
 - description: str
 - target_bom: ForeignKey(BillOfMaterials, backref='variants')
 - change_set: ForeignKey(VariantChangeSet), backref='included_in'
 - created_at: datetime
 - updated_at: datetime

### VariantChangeSet
 - id: pk
 - name: str
 - description: str
 - created_at: datetime
 - updated_at: datetime

### VariantChangeSetEntry
 - id: pk
 - change_set: ForeignKey(VariantChangeSet, backref='entries')
 - target_bom_entry: ForeignKey(BillOfMaterialsEntry), allow NULL for action:'ADD'
 - action:
    ```sql
    ENUM(
        'NO_POP',       -- Do not populate this entry from the base BOM
        'SUBSTITUTE',   -- Replace the target entry's part with a different one
        'ADD',          -- Add a new part to the BOM that is not in the base BOM
        'MODIFY_QTY',   -- Change the quantity of the target entry
        'NOTE_ONLY'     -- Add a note without changing the part or quantity
    )```
 - substitute_bom_entry: ForeignKey(BillOfMaterialsEntry), nullable

    - When target_bom_entry is NULL, we expect action to be 'ADD'. The resulting action is to add the substitute_bom_entry into the target BillOfMaterials.

 - notes: str
 - created_at: datetime
 - updated_at: datetime

#### TODO Resolved BOM Variant View (v_ResolvedBomVariant)

This view acts as a "materialized" BOM, presenting the final, flattened list of parts and quantities for a specific BillOfMaterialsVariant. It applies all ADD, SUBSTITUTE, MODIFY_QTY, and NO_POP rules, making it simple for the application layer to consume a variant BOM without complex logic.
