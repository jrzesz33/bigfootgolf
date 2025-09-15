# Pull Request Proposal: Fix EFS Tagging Issue in AWS Cloud Control MCP Server

## Problem

The AWS Cloud Control MCP Server has a bug where it always adds default tags using the `Tags` property, but some AWS resources use different tag property names. For example:

- **EFS (AWS::EFS::FileSystem)** uses `FileSystemTags` instead of `Tags`
- **RDS (AWS::RDS::DBInstance)** uses `Tags`  
- **S3 (AWS::S3::Bucket)** uses `Tags`

This causes validation errors when creating EFS resources:
```
AWS API Error (ValidationException): Model validation failed (#: extraneous key [Tags] is not permitted)
```

## Root Cause

In `src/ccapi-mcp-server/awslabs/ccapi_mcp_server/cloud_control_utils.py`, the `add_default_tags()` function always uses `Tags`:

```python
def add_default_tags(properties: Dict, schema: Dict) -> Dict:
    # ... existing code ...
    
    # Always uses 'Tags' regardless of resource type
    if 'Tags' not in modified_properties:
        modified_properties['Tags'] = []
    
    existing_tag_keys = [tag.get('Key') for tag in modified_properties.get('Tags', [])]
    # ... rest of function
```

## Solution

Create a resource-type-aware tagging system that maps AWS resource types to their correct tag property names.

### Proposed Changes

**1. Create Resource Type Mapping**

Add a mapping dictionary for resources that use non-standard tag properties:

```python
# Resource types that use non-standard tag properties
RESOURCE_TAG_PROPERTY_MAP = {
    'AWS::EFS::FileSystem': 'FileSystemTags',
    'AWS::EFS::AccessPoint': 'Tags',  # EFS Access Points use Tags
    'AWS::CloudFormation::Stack': 'Tags',
    'AWS::ECR::Repository': 'Tags',
    # Add more as discovered
    # Default fallback is 'Tags' for most AWS resources
}
```

**2. Update add_default_tags Function**

Modify the function signature and implementation:

```python
def add_default_tags(properties: Dict, schema: Dict, resource_type: str = None) -> Dict:
    """
    Add default MCP server tags to resource properties.
    
    Args:
        properties: The resource properties dictionary
        schema: The resource schema (for future validation)
        resource_type: AWS resource type (e.g., 'AWS::EFS::FileSystem')
    
    Returns:
        Modified properties dictionary with default tags added
    """
    if not properties:
        return {}
    
    # Create a copy to avoid modifying the original
    modified_properties = properties.copy()
    
    # Determine the correct tag property name for this resource type
    tag_property = RESOURCE_TAG_PROPERTY_MAP.get(resource_type, 'Tags')
    
    # Ensure the tag property exists
    if tag_property not in modified_properties:
        modified_properties[tag_property] = []
    
    # Get existing tag keys to avoid duplicates
    existing_tag_keys = [
        tag.get('Key') for tag in modified_properties.get(tag_property, [])
    ]
    
    # Define default tags
    default_tags = [
        {'Key': 'MANAGED_BY', 'Value': 'CCAPI-MCP-SERVER'},
        {'Key': 'MCP_SERVER_SOURCE_CODE', 
         'Value': 'https://github.com/awslabs/mcp/tree/main/src/ccapi-mcp-server'},
        {'Key': 'MCP_SERVER_VERSION', 'Value': get_package_version()}
    ]
    
    # Add missing default tags
    for default_tag in default_tags:
        if default_tag['Key'] not in existing_tag_keys:
            modified_properties[tag_property].append(default_tag)
    
    return modified_properties
```

**3. Update Function Calls**

Find all places where `add_default_tags()` is called and pass the resource type:

```python
# Before
properties = add_default_tags(properties, schema)

# After  
properties = add_default_tags(properties, schema, resource_type)
```

**4. Add Unit Tests**

Create tests to verify the fix works for different resource types:

```python
def test_add_default_tags_efs_resource():
    """Test that EFS resources get FileSystemTags instead of Tags"""
    properties = {"PerformanceMode": "generalPurpose"}
    result = add_default_tags(properties, {}, "AWS::EFS::FileSystem")
    
    assert "FileSystemTags" in result
    assert "Tags" not in result
    assert len(result["FileSystemTags"]) == 3  # Three default tags

def test_add_default_tags_s3_resource():
    """Test that S3 resources still get Tags (default behavior)"""
    properties = {"BucketName": "test-bucket"}
    result = add_default_tags(properties, {}, "AWS::S3::Bucket")
    
    assert "Tags" in result
    assert "FileSystemTags" not in result
    assert len(result["Tags"]) == 3
```

## Files to Modify

1. **`src/ccapi-mcp-server/awslabs/ccapi_mcp_server/cloud_control_utils.py`**
   - Add `RESOURCE_TAG_PROPERTY_MAP` constant
   - Update `add_default_tags()` function

2. **Find and update all callers of `add_default_tags()`** - likely in:
   - Resource creation handlers
   - Infrastructure generation code
   - Any file that imports from `cloud_control_utils`

3. **`tests/` directory**
   - Add unit tests for the new tagging logic
   - Test multiple resource types (EFS, S3, RDS, etc.)

## Benefits

1. **Fixes EFS Resources**: EFS file systems can now be created successfully
2. **Future-Proof**: Easy to add more resource types with non-standard tag properties
3. **Backwards Compatible**: Existing resources using `Tags` continue to work
4. **Maintainable**: Centralized mapping makes it easy to add new resource types

## Testing

1. Test EFS file system creation (currently fails)
2. Test S3 bucket creation (should continue working)
3. Test RDS instance creation (should continue working)
4. Run existing unit test suite to ensure no regressions

## Alternative Solutions Considered

1. **Remove automatic tagging entirely** - Not ideal as tags provide value for resource management
2. **Use schema inspection** - Complex and unreliable due to schema inconsistencies
3. **Try/catch approach** - Would hide real validation errors

The proposed solution provides the best balance of functionality, maintainability, and reliability.

---

**Repository**: https://github.com/awslabs/mcp/tree/main/src/ccapi-mcp-server  
**Issue**: EFS resources fail with "extraneous key [Tags] is not permitted"  
**Priority**: High (blocks EFS resource creation completely)