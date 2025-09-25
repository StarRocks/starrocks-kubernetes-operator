# StarRocks Operator Upgrade Hook Feature

This feature adds support for pre-upgrade hooks to the StarRocks Kubernetes Operator, allowing automated execution of SQL commands before image upgrades.

## Overview

The StarRocks Operator now supports:

- **Pre-upgrade hooks**: Execute SQL commands before upgrading images
- **Annotation-based triggers**: Use annotations to trigger upgrade preparation  
- **Spec-based configuration**: Define hooks in the StarRocksCluster spec
- **Status tracking**: Monitor upgrade preparation progress
- **Post-upgrade cleanup**: Automatically re-enable services after upgrade

## Files Modified/Added

1. **api_modifications.go** - API enhancements for CRD
2. **upgrade_hook_controller.go** - New controller for handling upgrade hooks
3. **main_controller_modifications.go** - Integration with main controller
4. **upgrade_hook_controller_test.go** - Unit tests
5. **USAGE_EXAMPLES.md** - Usage documentation
6. **Makefile** - Build configuration

## Integration Steps

1. **Update API definitions**: Add new types to the StarRocks CRD
2. **Add controller**: Integrate upgrade hook controller
3. **Modify main reconciler**: Add hook execution logic
4. **Update RBAC**: Ensure proper permissions
5. **Add dependencies**: Include MySQL driver for SQL execution

## Benefits

- **Automated upgrade preparation**: No manual SQL commands required
- **Consistent upgrades**: Ensures proper pre-upgrade steps are executed
- **Flexible configuration**: Support both annotation and spec-based configuration
- **Error handling**: Proper error handling and rollback capabilities
- **Status visibility**: Clear status reporting for upgrade preparation

## Dependencies

```go
require (
    github.com/go-sql-driver/mysql v1.7.1
    // existing dependencies...
)
```

## Next Steps

1. Review the implementation
2. Add to feature branch
3. Create comprehensive tests
4. Update documentation
5. Submit pull request
