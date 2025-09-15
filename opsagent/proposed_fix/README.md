# AWS Cloud Control MCP Server - EFS Tagging Fix

This directory contains the proposed fix for the EFS tagging issue in the AWS Cloud Control MCP Server.

## Problem Summary

The MCP server fails to create EFS resources because it always adds default tags using the `Tags` property, but EFS resources require `FileSystemTags` instead. This causes the error:

```
AWS API Error (ValidationException): Model validation failed (#: extraneous key [Tags] is not permitted)
```

## Solution Overview

The fix introduces resource-type-aware tagging by:

1. **Mapping resource types to their correct tag properties**
2. **Updating the `add_default_tags()` function** to use the correct property name
3. **Maintaining backward compatibility** for existing resources

## Files in this Fix

### `cloud_control_utils.py`
- **Fixed version** of the original file
- Adds `RESOURCE_TAG_PROPERTY_MAP` for resource-specific tag properties
- Updates `add_default_tags()` to accept and use `resource_type` parameter
- Includes helper functions for extending the mapping

### `test_cloud_control_utils.py`
- **Comprehensive unit tests** for the new tagging logic
- Tests EFS resources get `FileSystemTags`
- Tests S3 resources still get `Tags` (backward compatibility)
- Tests edge cases like unknown resource types and existing tags

### Key Changes Made

#### 1. Resource Type Mapping
```python
RESOURCE_TAG_PROPERTY_MAP = {
    'AWS::EFS::FileSystem': 'FileSystemTags',
    'AWS::EFS::AccessPoint': 'Tags',
    'AWS::S3::Bucket': 'Tags',
    'AWS::RDS::DBInstance': 'Tags',
    # ... more resource types
}
```

#### 2. Updated Function Signature
```python
# Before
def add_default_tags(properties: Dict, schema: Dict) -> Dict:

# After  
def add_default_tags(properties: Dict, schema: Dict, resource_type: str = None) -> Dict:
```

#### 3. Dynamic Tag Property Selection
```python
# Determine the correct tag property name for this resource type
tag_property = RESOURCE_TAG_PROPERTY_MAP.get(resource_type, 'Tags')

# Use the correct property
if tag_property not in modified_properties:
    modified_properties[tag_property] = []
```

## Testing the Fix

### Run Unit Tests
```bash
cd proposed_fix/
python -m pytest test_cloud_control_utils.py -v
```

### Expected Test Results
- ✅ EFS resources get `FileSystemTags` 
- ✅ S3 resources get `Tags` (backward compatibility)
- ✅ Unknown resources default to `Tags`
- ✅ Existing tags are preserved
- ✅ No duplicate tags are added

## Integration Steps

To integrate this fix into the main MCP server:

1. **Replace** `src/ccapi-mcp-server/awslabs/ccapi_mcp_server/cloud_control_utils.py`
2. **Update all callers** of `add_default_tags()` to pass the `resource_type` parameter
3. **Add unit tests** to the test suite
4. **Run integration tests** to ensure no regressions

### Example Caller Update
```python
# Before
properties = add_default_tags(properties, schema)

# After
properties = add_default_tags(properties, schema, resource_type)
```

## Verification

After applying the fix, you should be able to:

1. **Create EFS file systems** without validation errors
2. **Continue creating other resources** (S3, RDS, etc.) without issues  
3. **See proper tags** on all resources using their correct property names

## Future Extensions

New resource types with non-standard tag properties can be easily added to `RESOURCE_TAG_PROPERTY_MAP`:

```python
# Add support for hypothetical new service
RESOURCE_TAG_PROPERTY_MAP['AWS::NewService::Resource'] = 'CustomTags'
```

## Impact Analysis

- **✅ Fixes**: EFS resource creation
- **✅ Maintains**: Backward compatibility for all existing resources
- **✅ Enables**: Easy extension to other non-standard resources
- **✅ Improves**: Code maintainability and testability
- **❌ No breaking changes**: All existing functionality preserved

This fix resolves the immediate EFS issue while providing a robust foundation for handling similar tagging inconsistencies across AWS services.